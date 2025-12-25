package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"openserver/utils"
	"os"
	"os/exec"
	"strings"
)

type hostnameRequest struct {
	Hostname string `json:"hostname"`
	Service  string `json:"service"`
}

type jsonList struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	CreatedAt   string `json:"created_at"`
	DeletedAt   string `json:"deleted_at"`
	Connections string `json:"connections"`
}

func main() {
	command := `cloudflared tunnel list --output json`

	shCmd := exec.Command("sh", "-c", command)
	var stdOut, stdErr bytes.Buffer

	shCmd.Stderr = &stdErr
	shCmd.Stdout = &stdOut

	if err := shCmd.Run(); err != nil {
		utils.ErrorLogger.Printf("Command failed: %s\nError while getting tunnel-id: %s", err.Error(), stdErr.String())
		return
	}

	var tunnelJsonList []jsonList
	err := json.Unmarshal(stdOut.Bytes(), &tunnelJsonList)
	if err != nil {
		utils.ErrorLogger.Printf("Failed to parse tunnel list: %s", err.Error())
		return
	}

	if len(tunnelJsonList) == 0 {
		utils.ErrorLogger.Println("Empty tunnel list")
		return
	}
	tunnelId := ""
	for i := range tunnelJsonList {
		if tunnelJsonList[i].Name == "openserver-tunnel" {
			tunnelId = tunnelJsonList[i].Id
			break
		}
	}

	if tunnelId == "" {
		utils.ErrorLogger.Println("No tunnel named openserver-tunnel found")
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

		cmd := exec.Command("cloudflared", "tunnel", "route", "dns", tunnelId, request.Hostname)

		var stdErr bytes.Buffer
		cmd.Stderr = &stdErr

		if err = cmd.Run(); err != nil {
			utils.ErrorLogger.Printf("Error while routing dns: %s", stdErr.String())
			utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
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
		stdErr.Reset()
		stdOut.Reset()

		killCloudflared := exec.Command("pkill", "-f", `"cloudflared tunnel run"`)

		killCloudflared.Stderr = &stdErr
		killCloudflared.Stdout = &stdOut

		if err = killCloudflared.Run(); err != nil {
			utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			utils.ErrorLogger.Printf("Error while listing processes: %s", stdErr.String())
			return
		}

		if strings.Trim(stdOut.String(), " ") == "" {
			utils.ErrorLogger.Printf("Failed to kill the tunnel process")
			return
		}
		stdErr.Reset()
		stdOut.Reset()

		rerunTunnel := exec.Command("cloudflared", "tunnel", "run", "openserver-tunnel")
		rerunTunnel.Stderr = &stdErr

		if err = rerunTunnel.Run(); err != nil {
			utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			utils.ErrorLogger.Printf("Falied to start the tunnel: %s", stdErr.String())
			return
		}
	})

	http.ListenAndServe(":8080", nil)
}
