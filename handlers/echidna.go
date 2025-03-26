package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func RunEchidna(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		TestFilePath string `json:"testFilePath"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Failed to parse request")
		return
	}

	// Verify file exists
	if _, err := os.Stat(req.TestFilePath); os.IsNotExist(err) {
		sendErrorResponse(w, "Test file not found")
		return
	}

	// Create a temporary config file for Echidna
	configContent := `
corpusDir: "corpus"
testMode: "property"
testLimit: 50000
timeout: 300
seqLen: 100
shrinkLimit: 1000
coverage: true
format: "text"
`
	configPath := filepath.Join(filepath.Dir(req.TestFilePath), "echidna.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		sendErrorResponse(w, "Failed to create Echidna config file")
		return
	}

	// Prepare Echidna command
	cmd := exec.Command("echidna",
		req.TestFilePath,
		"--config", configPath,
		"--format", "text",
		"--contract", "EchidnaTest",
		"--corpus-dir", "corpus",
		"--test-mode", "property",
	)

	// Create buffer to capture output
	var outputBuffer bytes.Buffer
	cmd.Stdout = &outputBuffer
	cmd.Stderr = &outputBuffer

	// Set timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd = exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)

	// Run Echidna
	err := cmd.Run()

	// Process output
	output := outputBuffer.String()

	// Check for timeout
	if ctx.Err() == context.DeadlineExceeded {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Echidna execution timed out after 5 minutes",
			"results": output,
		})
		return
	}

	// Check for other errors
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Echidna exited with code %d", exitErr.ExitCode()),
				"results": output,
			})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Error running Echidna: %v", err),
			"results": output,
		})
		return
	}

	// Parse and format the results
	results := parseEchidnaOutput(output)

	// Return results
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"results": results,
	})

	// Cleanup
	os.Remove(configPath)
}

func parseEchidnaOutput(output string) map[string]interface{} {
	results := make(map[string]interface{})

	lines := strings.Split(output, "\n")

	var failedTests []string
	var passedTests []string
	var coverage string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "failed!") {
			failedTests = append(failedTests, line)
		}

		if strings.Contains(line, "passed!") {
			passedTests = append(passedTests, line)
		}

		if strings.Contains(line, "coverage:") {
			coverage = line
		}
	}

	results["failed_tests"] = failedTests
	results["passed_tests"] = passedTests
	results["coverage"] = coverage
	results["raw_output"] = output

	return results
}
