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
	code := r.FormValue("solidityCode")
	if code == "" {
		http.Error(w, "No code provided", http.StatusBadRequest)
		return
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		http.Error(w, "OpenAI API key not set", http.StatusInternalServerError)
		return
	}

	// Create the request body for chat completions
	reqBody := OpenAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: fmt.Sprintf("Analyze this Solidity smart contract and list potential invariants. Return each invariant on a new line:\n\n%s", code),
			},
		},
		Temperature: 0.7,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		http.Error(w, "Error preparing request", http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(reqJSON))
	if err != nil {
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error calling OpenAI API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		http.Error(w, "Error decoding OpenAI response", http.StatusInternalServerError)
		return
	}

	// Process the response into individual invariants
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

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"invariants": invariants,
	})
}
