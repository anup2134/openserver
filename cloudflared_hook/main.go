package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"openserver/utils"
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

	var tunnelJSONList []jsonList
	err := json.Unmarshal(stdOut.Bytes(), &tunnelJSONList)
	if err != nil {
		utils.ErrorLogger.Printf("Failed to parse tunnel list: %s", err.Error())
		return
	}

	if len(tunnelJSONList) == 0 {
		utils.ErrorLogger.Println("Empty tunnel list")
		return
	}
	tunnelID := ""
	for i := range tunnelJSONList {
		if tunnelJSONList[i].Name == "openserver-tunnel" {
			tunnelID = tunnelJSONList[i].Id
			break
		}
	}

	if tunnelID == "" {
		utils.ErrorLogger.Println("No tunnel named openserver-tunnel found")
		return
	}

	http.HandleFunc("/add_hostname", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.SendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		defer utils.CloseRequestBody(r)
		request := &hostnameRequest{}
		err := utils.DecodeRequestJSON(r.Body, request)
		if err != nil {
			utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}

		cmd := exec.Command("cloudflared", "tunnel", "route", "dns", tunnelID, request.Hostname)

		var stdErr bytes.Buffer
		cmd.Stderr = &stdErr

		if err = cmd.Run(); err != nil {
			utils.ErrorLogger.Printf("Error while routing dns: %s", stdErr.String())
			utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		content := fmt.Sprintf("  - hostname: %s\n    service: %s\n  - service: http_status:404\n\n", request.Hostname, request.Service)
		f, err := os.OpenFile("/home/nonroot/.cloudflared/config.yml", os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() {
			err := f.Close()
			if err != nil {
				utils.ErrorLogger.Printf("Failed to close config.yml file: %s\n", err.Error())
			}
		}()

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

		if strings.TrimSpace(stdOut.String()) == "" {
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

	http.HandleFunc("/get_available_port", func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile("/home/nonroot/.cloudflared/Records.txt")
		if err != nil {
			utils.ErrorLogger.Printf("Failed to read Records.txt file: %s\n", err.Error())
			utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		records := strings.Split(string(data), "\n")
		if len(records) == 0 {
			_, err := w.Write([]byte("3000"))
			if err != nil {
				utils.ErrorLogger.Printf("Failed to write to response writer: %s\n", err.Error())
				utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}
		lastRecord := strings.Split(records[len(records)-1], ",")

		lastUsedPort, err := strconv.Atoi(strings.TrimSpace(lastRecord[2]))
		if err != nil {
			utils.ErrorLogger.Printf("Invalid file format for Records.txt: %s", err.Error())
			utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		nextPort := strconv.FormatUint(uint64(lastUsedPort)+1, 10)
		_, err = w.Write([]byte(nextPort))
		if err != nil {
			utils.ErrorLogger.Printf("Failed to write to response writer: %s\n", err.Error())
			utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		utils.ErrorLogger.Printf("Failed to start the http server: %s", err.Error())
	}
}
