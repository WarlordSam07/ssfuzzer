package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type TestGenerationRequest struct {
	SolidityCode string   `json:"solidityCode"`
	Invariants   []string `json:"invariants"`
}

type TestGenerationResponse struct {
	Success     bool   `json:"success"`
	TestCode    string `json:"testCode,omitempty"`
	Error       string `json:"error,omitempty"`
	EchidnaFile string `json:"echidnaFile,omitempty"`
}

type OpenAIRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func GenerateTests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request
	var req TestGenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Failed to parse request body")
		return
	}

	// Validate input
	if strings.TrimSpace(req.SolidityCode) == "" {
		sendErrorResponse(w, "No Solidity code provided")
		return
	}
	if len(req.Invariants) == 0 {
		sendErrorResponse(w, "No invariants selected")
		return
	}

	// Prepare prompt for test generation
	prompt := fmt.Sprintf(`Generate a Solidity test contract that uses Echidna to test the following invariants:

%s

For this contract:

%s

The test contract should:
1. Inherit from the original contract
2. Include property functions for each invariant
3. Follow Echidna testing conventions
4. Include necessary setup in the constructor
5. Return only the complete test contract code`,
		strings.Join(req.Invariants, "\n"), req.SolidityCode)

	// Create OpenAI request
	openAIReqBody := OpenAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: "You are a Solidity smart contract test generator specializing in Echidna fuzzing tests.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
	}

	// Get API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		sendErrorResponse(w, "OpenAI API key not configured")
		return
	}

	// Send request to OpenAI
	reqJSON, err := json.Marshal(openAIReqBody)
	if err != nil {
		sendErrorResponse(w, "Error preparing request")
		return
	}

	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(reqJSON))
	if err != nil {
		sendErrorResponse(w, "Error creating request")
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: time.Second * 30}
	resp, err := client.Do(httpReq)
	if err != nil {
		sendErrorResponse(w, "Error connecting to OpenAI")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		sendErrorResponse(w, fmt.Sprintf("OpenAI API error (status %d): %s", resp.StatusCode, string(body)))
		return
	}

	// Parse OpenAI response
	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		sendErrorResponse(w, "Error parsing OpenAI response")
		return
	}

	if len(openAIResp.Choices) == 0 {
		sendErrorResponse(w, "No response from OpenAI")
		return
	}

	testCode := openAIResp.Choices[0].Message.Content

	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "echidna_test_*")
	if err != nil {
		sendErrorResponse(w, "Error creating temporary directory")
		return
	}

	// Save test file
	testFilePath := filepath.Join(tempDir, "EchidnaTest.sol")
	if err := os.WriteFile(testFilePath, []byte(testCode), 0644); err != nil {
		sendErrorResponse(w, "Error saving test file")
		return
	}

	// Return success response with test code
	json.NewEncoder(w).Encode(TestGenerationResponse{
		Success:     true,
		TestCode:    testCode,
		EchidnaFile: testFilePath,
	})
}

func sendErrorResponse(w http.ResponseWriter, message string) {
	json.NewEncoder(w).Encode(TestGenerationResponse{
		Success: false,
		Error:   message,
	})
}
