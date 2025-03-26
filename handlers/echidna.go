package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func RunEchidna(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req EchidnaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Failed to parse request body")
		return
	}

	if strings.TrimSpace(req.TestCode) == "" {
		sendErrorResponse(w, "No test code provided")
		return
	}

	// Create temporary directory
	tmpDir, err := ioutil.TempDir("", "echidna_test_")
	if err != nil {
		sendErrorResponse(w, "Failed to create temporary directory")
		return
	}
	defer os.RemoveAll(tmpDir)

	// Create Echidna config file
	configContent := `
testLimit: 50000
coverage: true
corpusDir: "corpus"
testMode: "property"
cryticArgs: ["--solc-version", "0.8.0"]
filterFunctions: ["echidna"]
seqLen: 100
shrinkLimit: 5000
timeout: 300
`
	configPath := filepath.Join(tmpDir, "echidna-config.yaml")
	if err := ioutil.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		sendErrorResponse(w, "Failed to create config file")
		return
	}

	// Save test file
	testFilePath := filepath.Join(tmpDir, "EchidnaTest.sol")
	if err := ioutil.WriteFile(testFilePath, []byte(req.TestCode), 0644); err != nil {
		sendErrorResponse(w, "Failed to write test file")
		return
	}

	// Run Echidna with improved configuration
	cmd := exec.Command("echidna",
		testFilePath,
		"--config", configPath,
		"--format", "text",
		"--contract", "EchidnaTest")

	output, err := cmd.CombinedOutput()

	// Process output
	response := EchidnaResponse{
		Success: true,
		Output:  string(output),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			response.Success = false
			response.Error = fmt.Sprintf("Echidna exited with code %d", exitErr.ExitCode())
		} else {
			sendErrorResponse(w, "Failed to run Echidna")
			return
		}
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		sendErrorResponse(w, "Failed to encode response")
		return
	}
}
