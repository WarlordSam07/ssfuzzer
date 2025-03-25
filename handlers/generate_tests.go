package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func GenerateTests(w http.ResponseWriter, r *http.Request) {
	var request struct {
		SolidityCode string   `json:"solidityCode"`
		Invariants   []string `json:"invariants"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Failed to parse request", http.StatusBadRequest)
		return
	}

	// Generate Echidna test contract
	var testContract strings.Builder
	testContract.WriteString("contract TestContract is ")
	testContract.WriteString(extractContractName(request.SolidityCode))
	testContract.WriteString(" {\n")

	// Add invariant functions
	for i, invariant := range request.Invariants {
		testContract.WriteString(fmt.Sprintf(`
    function echidna_test_%d() public returns (bool) {
        // %s
        return true; // Replace with actual invariant check
    }
`, i+1, invariant))
	}

	testContract.WriteString("}\n")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"testFiles": testContract.String(),
	})
}

func extractContractName(code string) string {
	// Simple extraction of contract name - could be improved
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		if strings.Contains(line, "contract") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "contract" && i+1 < len(parts) {
					return strings.TrimSpace(parts[i+1])
				}
			}
		}
	}
	return "UnknownContract"
}
