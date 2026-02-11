package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"openserver/utils"
)

type buildRequest struct {
	Owner      string `json:"owner"`
	Event      string `json:"event"`
	RepoName   string `json:"repoName"`
	CommitHash string `json:"commit"`
}

var tmpCloneDir string = "/home/nonroot/cloneTmp/"

func main() {
	runtime.GOMAXPROCS(1)

	http.HandleFunc("/build_image", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.SendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		defer closeRequestBody(r)

		request := &buildRequest{}
		err := utils.DecodeRequestJSON(r.Body, request)
		if err != nil {
			utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		if request.Event == "" || request.RepoName == "" || request.Owner == "" || request.CommitHash == "" {
			utils.SendErrorResponse(w, "missing required fields", http.StatusBadRequest)
			return
		}

		utils.InfoLogger.Printf("Payload:\n{\n\tevent: %s,\n\towner: %s,\n\trepoName: %s,\n\tcommit: %s\n}\n", request.Event, request.Owner, request.RepoName, request.CommitHash)

		authHeader := r.Header.Get("Authorization")
		err = authorization(authHeader)
		if err != nil {
			utils.SendErrorResponse(w, err.Error(), http.StatusUnauthorized)
			return
		}
		cloneDir := tmpCloneDir + request.CommitHash
		gitCmd := exec.Command("git", "clone", "https://github.com/"+request.Owner+"/"+request.RepoName, cloneDir)

		var stdOut, stdErr bytes.Buffer

		gitCmd.Stderr = &stdErr
		gitCmd.Stdout = &stdOut

		if err := gitCmd.Run(); err != nil {
			utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			utils.ErrorLogger.Printf("Error while cloning repo: %s", stdErr.String())
			return
		}
		defer cleanupCloneDir(cloneDir)
		imageTag := strings.ToLower(request.Owner + "." + request.RepoName)

		err = buildDockerImage(cloneDir, imageTag)
		if err != nil {
			utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Error occured while starting http server: %s\n", err.Error())
	}
}

func authorization(authHeader string) error {
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		utils.ErrorLogger.Println("invalid auth header")
		return fmt.Errorf("invalid auth header")
	}

	authToken := authHeader[7:]
	if authToken != os.Getenv("secretKey") {
		utils.ErrorLogger.Println("incorrect auth token")
		return fmt.Errorf("incorrect auth token")
	}
	return nil
}

func cleanupCloneDir(dirName string) {
	err := os.RemoveAll(tmpCloneDir + dirName)
	if err != nil {
		utils.ErrorLogger.Printf("Error while deleting clone directory: %s", err.Error())
	}
}

func closeRequestBody(r *http.Request) {
	err := r.Body.Close()
	if err != nil {
		utils.ErrorLogger.Printf("Error occured while closing request body: %s\n", err.Error())
	}
}

func buildDockerImage(pathToCloneDir string, imageTag string) error {
	dockerCmd := exec.Command("docker", "build", "-t", imageTag, pathToCloneDir)

	var stdOut, stdErr bytes.Buffer
	dockerCmd.Stderr = &stdErr
	dockerCmd.Stdout = &stdOut

	if err := dockerCmd.Run(); err != nil {
		dockerErrorMessage := stdErr.String()
		utils.ErrorLogger.Printf("Docker build failed: %v | Command Output: %s\n", err, dockerErrorMessage)
		return fmt.Errorf("docker build error: %w (details: %s)", err, dockerErrorMessage)
	}

	return nil
}
