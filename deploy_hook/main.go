package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
)

type BuildRequest struct {
	Owner      string `json:"owner"`
	Event      string `json:"event"`
	RepoName   string `json:"repoName"`
	CommitHash string `json:"commit"`
}

var (
	InfoLogger  = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
)

func SendErrorResponse(w http.ResponseWriter, msg string, code int) {
	InfoLogger.Print(msg)
	http.Error(w, msg, code)
}

func main() {
	runtime.GOMAXPROCS(1)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()
		buildRequest, err := DecodeRequestJSON(r.Body)
		if err != nil {
			SendErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		InfoLogger.Printf("Payload:\n{\n\tevent: %s,\n\towner: %s,\n\trepoName: %s,\n\tcommit: %s\n}\n", buildRequest.Event, buildRequest.Owner, buildRequest.RepoName, buildRequest.CommitHash)

		authHeader := r.Header.Get("Authorization")
		err = authorization(authHeader)
		if err != nil {
			SendErrorResponse(w, err.Error(), http.StatusUnauthorized)
			return
		}
	})

	http.ListenAndServe(":8080", nil)
}

func DecodeRequestJSON(requestBody io.ReadCloser) (BuildRequest, error) {
	var buildRequest BuildRequest
	dec := json.NewDecoder(requestBody)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&buildRequest); err != nil {
		ErrorLogger.Printf("invalid JSON: %s\n", err.Error())
		return BuildRequest{RepoName: "", Event: "", CommitHash: ""}, fmt.Errorf("invalid JSON: %s", err.Error())
	}
	if buildRequest.Event == "" || buildRequest.RepoName == "" || buildRequest.Owner == "" || buildRequest.CommitHash == "" {
		ErrorLogger.Printf("missing required fields\n")
		return BuildRequest{RepoName: "", Event: "", CommitHash: ""}, fmt.Errorf("missing required fields")
	}

	return buildRequest, nil
}

func authorization(authHeader string) error {
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		ErrorLogger.Println("invalid auth header")
		return fmt.Errorf("invalid auth header")
	}

	authToken := authHeader[7:]
	if authToken != os.Getenv("secretKey") {
		ErrorLogger.Println("incorrect auth token")
		return fmt.Errorf("incorrect auth token")
	}
	return nil
}
