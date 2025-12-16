package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var (
	InfoLogger  = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
)

func SendErrorResponse(w http.ResponseWriter, msg string, code int) {
	InfoLogger.Print(msg)
	http.Error(w, msg, code)
}

func DecodeRequestJSON(requestBody io.ReadCloser, jsonSchema any) error {
	dec := json.NewDecoder(requestBody)
	dec.DisallowUnknownFields()

	if err := dec.Decode(jsonSchema); err != nil {
		ErrorLogger.Printf("Invalid JSON: %s\n", err.Error())
		return fmt.Errorf("invalid JSON")
	}
	return nil
}
