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

type EchidnaRequest struct {
	TestCode string `json:"testCode"`
}

type EchidnaResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

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

	// Create a temporary directory for the test files
	tmpDir, err := ioutil.TempDir("", "echidna_test_")
	if err != nil {
		sendErrorResponse(w, "Failed to create temporary directory")
		return
	}
	defer os.RemoveAll(tmpDir)

	// Create the test file
	testFilePath := filepath.Join(tmpDir, "Test.sol")
	err = ioutil.WriteFile(testFilePath, []byte(req.TestCode), 0644)
	if err != nil {
		sendErrorResponse(w, "Failed to write test file")
		return
	}

	// Run Echidna
	cmd := exec.Command("echidna", testFilePath, "--config", "echidna-config.yaml")
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Check if it's just a non-zero exit code (which Echidna might return for failed tests)
		if _, ok := err.(*exec.ExitError); !ok {
			fmt.Printf("Error running Echidna: %v\n", err)
			sendErrorResponse(w, "Failed to run Echidna")
			return
		}
	}

	// Send the response
	response := EchidnaResponse{
		Success: true,
		Output:  string(output),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding response: %v\n", err)
		sendErrorResponse(w, "Failed to encode response")
		return
	}
}
