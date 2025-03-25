package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
)

// RunEchidna runs the Echidna fuzzing tool on the generated test files
func RunEchidna(w http.ResponseWriter, r *http.Request) {
	// Run Echidna tool (assuming test.sol is the generated test file)
	cmd := exec.Command("echidna", "test.sol") // Modify as necessary
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		http.Error(w, "Failed to run Echidna", http.StatusInternalServerError)
		return
	}

	// Return the results of the Echidna run
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": out.String(),
	})
}
