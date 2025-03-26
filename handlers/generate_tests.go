package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func GenerateTests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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

	// Extract contract name from the Solidity code
	contractName := extractContractName(req.SolidityCode)
	if contractName == "" {
		contractName = "UnnamedContract"
	}

	// Create directory with timestamp
	timestamp := time.Now().Format("20060102150405")
	testDir := filepath.Join(".", "tests", contractName+"_"+timestamp)
	if err := os.MkdirAll(testDir, 0755); err != nil {
		sendErrorResponse(w, "Error creating test directory")
		return
	}

	// Save original contract
	contractPath := filepath.Join(testDir, "Contract.sol")
	if err := os.WriteFile(contractPath, []byte(req.SolidityCode), 0644); err != nil {
		sendErrorResponse(w, "Error saving contract file")
		return
	}

	var testFiles []string
	var testContents []string

	// Generate separate test for each selected invariant
	for i, invariant := range req.Invariants {
		// Create test name based on invariant number
		testFunctionName := fmt.Sprintf("invariant_%d", i+1)

		prompt := fmt.Sprintf(`Generate a Solidity test contract for Echidna fuzzing that follows these exact specifications:

Contract to test:
%s

Invariant to verify:
%s

Requirements:
1. Use this exact format:
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./Contract.sol";

contract EchidnaTest_%d is Contract {
    constructor() Contract() {
        // Initialize state here with meaningful values
        // Set up any required initial conditions
    }
    
    function echidna_%s() public returns (bool) {
        // Testing invariant: %s
        // IMPLEMENT THE FOLLOWING LOGIC:
        // 1. Set up any pre-conditions needed
        // 2. Execute the relevant contract operations
        // 3. Check the invariant condition
        // 4. Return true if the invariant holds, false if violated
    }
}`,
			req.SolidityCode,
			invariant,
			i+1,
			testFunctionName,
			invariant)

		// Create OpenAI request
		openAIReqBody := OpenAIRequest{
			Model: "gpt-4",
			Messages: []ChatMessage{
				{
					Role:    "system",
					Content: "You are an expert Solidity test generator specializing in Echidna property-based fuzzing tests.",
				},
				{
					Role:    "user",
					Content: prompt,
				},
			},
			Temperature: 0.2,
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

		// Save individual test file
		testFileName := fmt.Sprintf("invariant_%d_test.sol", i+1)
		testFilePath := filepath.Join(testDir, testFileName)
		if err := os.WriteFile(testFilePath, []byte(testCode), 0644); err != nil {
			sendErrorResponse(w, fmt.Sprintf("Error saving test file: %s", testFileName))
			return
		}

		testFiles = append(testFiles, testFilePath)
		testContents = append(testContents, testCode)
	}

	// Return success response with all generated test files and their contents
	json.NewEncoder(w).Encode(TestGenerationResponse{
		Success:      true,
		TestCode:     req.SolidityCode,
		EchidnaFile:  testDir,
		TestFiles:    testFiles,
		TestContents: testContents,
	})
}

func extractContractName(code string) string {
	re := regexp.MustCompile(`contract\s+(\w+)`)
	matches := re.FindStringSubmatch(code)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
