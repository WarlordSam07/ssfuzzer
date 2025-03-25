package handlers

import (
	"encoding/json"
	"net/http"
	"os/exec"
)

func RunEchidna(w http.ResponseWriter, r *http.Request) {
	// Run Echidna command
	cmd := exec.Command("echidna-test", "TestContract.sol", "--config", "echidna.config.yml")
	output, err := cmd.CombinedOutput()

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Error running Echidna: " + err.Error() + "\n" + string(output),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": string(output),
	})
}
