package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

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

func SubmitCode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var reqBody struct {
		SolidityCode string `json:"solidityCode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to parse request body",
		})
		return
	}

	if reqBody.SolidityCode == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "No code provided",
		})
		return
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "OpenAI API key not set",
		})
		return
	}

	fmt.Printf("Sending the following code to OpenAI: \n%s\n", reqBody.SolidityCode)

	// Create the request body for chat completions
	openAIReqBody := OpenAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: fmt.Sprintf("Analyze this Solidity smart contract and list potential invariants. Return each invariant on a new line:\n\n%s", reqBody.SolidityCode),
			},
		},
		Temperature: 0.7,
	}

	reqJSON, err := json.Marshal(openAIReqBody)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Error preparing request",
		})
		return
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(reqJSON))
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Error creating request",
		})
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Error making request to OpenAI",
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("OpenAI API returned status code: %d", resp.StatusCode),
		})
		return
	}

	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Error parsing OpenAI response",
		})
		return
	}

	// Process the response and extract invariants
	// Process the response and extract invariants
	var invariants []string
	if len(openAIResp.Choices) > 0 {
		content := openAIResp.Choices[0].Message.Content
		lines := strings.Split(strings.TrimSpace(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				invariants = append(invariants, line)
			}
		}
	}

	// Log the response for debugging
	fmt.Printf("OpenAI Response: %+v\n", openAIResp)
	fmt.Printf("Processed invariants: %+v\n", invariants)

	// Send the final response
	response := map[string]interface{}{
		"success":    true,
		"invariants": invariants, // Make sure this is an array of strings
	}

	// Send the response in the expected format
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding response: %v\n", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Error encoding response",
		})
		return
	}

}
