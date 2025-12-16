package main

import (
	"net/http"
	"openserver/utils"
)

type hostnameRequest struct {
	Hostname string `json:"hostname"`
	Service  string `json:"service"`
}

func main() {
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

	})
}
