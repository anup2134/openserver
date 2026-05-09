package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"openserver/utils"
)

const (
	cloudflaredDir = "/home/nonroot/.cloudflared"
	configPath     = cloudflaredDir + "/config.yml"
)

type hostnameRequest struct {
	Hostname string `json:"hostname"`
	Service  string `json:"service"`
}

type ingressRule struct {
	Hostname string `yaml:"hostname,omitempty"`
	Service  string `yaml:"service"`
}

type cloudflaredConfig struct {
	Tunnel          string        `yaml:"tunnel,omitempty"`
	CredentialsFile string        `yaml:"credentials-file,omitempty"`
	Ingress         []ingressRule `yaml:"ingress"`
}

func addHostName(w http.ResponseWriter, r *http.Request) {
	tunnelID, err := getTunnelID()
	if err != nil {
		panic(err)
	}
	if r.Method != http.MethodPost {
		utils.SendErrorResponse(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer utils.CloseRequestBody(r)

	request := &hostnameRequest{}
	if err := utils.DecodeRequestJSON(r.Body, request); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if request.Hostname == "" || request.Service == "" {
		utils.SendErrorResponse(w, "hostname and service are required", http.StatusBadRequest)
		return
	}

	cmd := exec.Command(
		"cloudflared",
		"tunnel",
		"route",
		"dns",
		tunnelID,
		request.Hostname,
	)

	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr

	if err := cmd.Run(); err != nil {
		utils.ErrorLogger.Printf(
			"failed to route dns for hostname=%s: %s",
			request.Hostname,
			stdErr.String(),
		)

		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	config, err := readConfig(tunnelID)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	config.Ingress = append(
		config.Ingress,
		ingressRule{
			Hostname: request.Hostname,
			Service:  request.Service,
		},
	)

	ensure404Ingress(&config)

	if err := writeConfig(config); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write([]byte("success")); err != nil {
		utils.ErrorLogger.Printf(
			"failed to write response: %s",
			err.Error(),
		)
	}
}

func getTunnelID() (string, error) {
	files, err := os.ReadDir(cloudflaredDir)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()

		if strings.HasSuffix(name, ".json") {
			return strings.TrimSuffix(name, ".json"), nil
		}
	}

	return "", errors.New("no tunnel credentials json file found")
}

func readConfig(tunnelID string) (cloudflaredConfig, error) {
	config := cloudflaredConfig{
		Tunnel:          tunnelID,
		CredentialsFile: fmt.Sprintf("%s/%s.json", cloudflaredDir, tunnelID),
		Ingress:         []ingressRule{},
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := writeConfig(config); err != nil {
				return config, err
			}

			return config, nil
		}

		return config, err
	}

	if len(bytes.TrimSpace(data)) == 0 {
		if err := writeConfig(config); err != nil {
			return config, err
		}

		return config, nil
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, err
	}

	if config.Tunnel == "" {
		config.Tunnel = tunnelID
	}

	if config.CredentialsFile == "" {
		config.CredentialsFile = fmt.Sprintf(
			"%s/%s.json",
			cloudflaredDir,
			tunnelID,
		)
	}

	return config, nil
}

func writeConfig(config cloudflaredConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	tempFile := filepath.Join(cloudflaredDir, "config.yml.tmp")

	if err := os.WriteFile(tempFile, data, 0o644); err != nil {
		return err
	}

	return os.Rename(tempFile, configPath)
}

func ensure404Ingress(config *cloudflaredConfig) {
	for _, rule := range config.Ingress {
		if rule.Service == "http_status:404" {
			return
		}
	}

	config.Ingress = append(
		config.Ingress,
		ingressRule{
			Service: "http_status:404",
		},
	)
}
