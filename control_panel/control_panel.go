package main

import (
	"bytes"
	"fmt"
	"io"
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
	HostName   string `json:"hostname"` // domian/subdomain name for cloudflared tunnel config
}

var tmpCloneDir string = "/home/nonroot/cloneTmp/"

func main() {
	runtime.GOMAXPROCS(1)

	http.HandleFunc("/build_image", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.SendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		defer utils.CloseRequestBody(r)

		request := &buildRequest{}
		err := utils.DecodeRequestJSON(r.Body, request)
		if err != nil {
			utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		if request.Event == "" || request.RepoName == "" || request.Owner == "" || request.CommitHash == "" || request.HostName == "" {
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

		err = runDockerContainer(imageTag)
		if err != nil {
			deleteDockerImage(imageTag)
			utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cloudflaredRequestJSON := []byte(fmt.Sprintf(`{"hostname":"%s","service":"%s"}`, request.HostName, imageTag))
		req, err := http.NewRequest("POST", "http://cloudflared-container:8080/add_hostname", bytes.NewBuffer(cloudflaredRequestJSON))
		if err != nil {
			deleteDockerImage(imageTag)
			utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			utils.ErrorLogger.Printf("Failed to add hostname\nError Message: %s\n", err.Error())
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			deleteDockerImage(imageTag)
			utils.ErrorLogger.Printf("Failed to get the response\nError Message: %s\n", err.Error())
			utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if resp.StatusCode != 201 {
			deleteDockerImage(imageTag)
			utils.ErrorLogger.Printf("Failed to create record in cloudflared config\n")
			utils.SendErrorResponse(w, "Failed to create cloudflared config record", http.StatusInternalServerError)
			return
		}

		defer resp.Body.Close()

		_, err = io.ReadAll(resp.Body)
		if err != nil {
			deleteDockerImage(imageTag)
			utils.ErrorLogger.Printf("Failed to read response body\nError Message: %s\n", err.Error())
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

func runDockerContainer(imageTag string) error {
	dockerRunCmd := exec.Command("docker", "run", imageTag)

	var stdOut, stdErr bytes.Buffer
	dockerRunCmd.Stderr = &stdErr
	dockerRunCmd.Stdout = &stdOut
	if err := dockerRunCmd.Run(); err != nil {
		utils.ErrorLogger.Printf("Running docker container failed: %s\nCommand output: %s", err.Error(), stdErr.String())
		return err
	}
	return nil
}

func deleteDockerImage(imageTag string) {
	dockerImageRm := exec.Command("docker", "rmi", imageTag)

	var stdOut, stdErr bytes.Buffer
	dockerImageRm.Stderr = &stdErr
	dockerImageRm.Stdout = &stdOut

	if err := dockerImageRm.Run(); err != nil {
		utils.ErrorLogger.Printf("Failed to delete docker image%s\nCommand output: %s\n", imageTag, stdErr.String())
	}
}
