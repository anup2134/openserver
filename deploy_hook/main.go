package main

import (
	"fmt"
	"net/http"
	"openserver/utils"
	"os"
	"runtime"
)

type buildRequest struct {
	Owner      string `json:"owner"`
	Event      string `json:"event"`
	RepoName   string `json:"repoName"`
	CommitHash string `json:"commit"`
}

func main() {
	runtime.GOMAXPROCS(1)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.SendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()

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
	})

	http.ListenAndServe(":8080", nil)
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
