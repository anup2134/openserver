package main

import (
	"bytes"
	"fmt"
	"net/http"
	"openserver/utils"
	"os"
	"os/exec"
)

type hostnameRequest struct {
	Hostname string `json:"hostname"`
	Service  string `json:"service"`
}

func main() {
	command := `cloudflared tunnel list --output json | jq -r '.[] | select(.name=="openserver-tunnel") | .id'`

	cmd := exec.Command("bash", "-c", command)
	var stdIn, stdErr bytes.Buffer

	cmd.Stderr = &stdErr
	cmd.Stdin = &stdIn

	if err := cmd.Run(); err != nil {
		utils.ErrorLogger.Printf("Command failed: %s\nError while getting tunnel-id: %s", err.Error(), stdErr.String())
		return
	}

	tunnelId := stdIn.String()

	if tunnelId == "" {
		utils.ErrorLogger.Printf("No tunnel found named openserver-tunnel")
		return
	}

	http.HandleFunc("/add_hostname", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.SendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()
		request := &hostnameRequest{}
		err := utils.DecodeRequestJSON(r.Body, request)
		if err != nil {
			utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}

		content := fmt.Sprintf("  - hostname: %s\n    service: %s\n  - service: http_status:404\n\n", request.Hostname, request.Service)
		f, err := os.OpenFile("/home/nonroot/.cloudflared/config.yml", os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer f.Close()
		if _, err = f.WriteString(content); err != nil {
			utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cmd := exec.Command("cloudflared", "tunnel", "route", "dns", tunnelId, request.Hostname)
		var stdErr bytes.Buffer

		cmd.Stderr = &stdErr

		if err = cmd.Run(); err != nil {
			utils.ErrorLogger.Printf("Error while routing dns: %s", stdErr.String())
			utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

	})

	http.ListenAndServe(":8080", nil)
}
