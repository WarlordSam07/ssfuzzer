package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
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

type Response struct {
	Success    bool     `json:"success"`
	Invariants []string `json:"invariants,omitempty"`
	Error      string   `json:"error,omitempty"`
}

func SubmitCode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var reqBody struct {
		SolidityCode string `json:"solidityCode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		sendErrorResponse(w, "Failed to parse request body")
		return
	}

	if reqBody.SolidityCode == "" {
		sendErrorResponse(w, "No code provided")
		return
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		sendErrorResponse(w, "OpenAI API key not set")
		return
	}

	fmt.Printf("Processing Solidity code:\n%s\n", reqBody.SolidityCode)

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
		fmt.Printf("Error marshaling request: %v\n", err)
		sendErrorResponse(w, "Error preparing request")
		return
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(reqJSON))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		sendErrorResponse(w, "Error creating request")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: time.Second * 30}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		sendErrorResponse(w, "Error making request to OpenAI")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Printf("OpenAI API Error Response: %s\n", string(bodyBytes))
		sendErrorResponse(w, fmt.Sprintf("OpenAI API returned status code: %d", resp.StatusCode))
		return
	}

	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		sendErrorResponse(w, "Error parsing OpenAI response")
		return
	}

	var invariants []string
	if len(openAIResp.Choices) > 0 {
		content := openAIResp.Choices[0].Message.Content
		lines := strings.Split(strings.TrimSpace(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			line = strings.TrimPrefix(line, "- ")
			line = strings.TrimPrefix(line, "* ")
			if line != "" {
				invariants = append(invariants, line)
			}
		}
	}

	fmt.Printf("Processed invariants: %+v\n", invariants)

	response := Response{
		Success:    true,
		Invariants: invariants,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding response: %v\n", err)
		sendErrorResponse(w, "Error encoding response")
		return
	}
}

func sendErrorResponse(w http.ResponseWriter, message string) {
	response := Response{
		Success: false,
		Error:   message,
	}
	json.NewEncoder(w).Encode(response)
}
