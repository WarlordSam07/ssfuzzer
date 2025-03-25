package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

// Struct to handle the response from OpenAI
type OpenAIResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

// SubmitCode receives Solidity code and sends it to ChatGPT to detect potential invariants
func SubmitCode(w http.ResponseWriter, r *http.Request) {
	// Read the Solidity code from the request
	code := r.FormValue("solidityCode")

	// Log the code for debugging purposes
	log.Println("Received Solidity Code: ", code)

	// Get the OpenAI API key from environment variables
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		http.Error(w, "OpenAI API key not set", http.StatusInternalServerError)
		return
	}

	// Prepare the prompt for ChatGPT
	prompt := fmt.Sprintf("Analyze the following Solidity code and detect potential invariants:\n\n%s", code)

	// Send the request to OpenAI API
	client := &http.Client{}
	reqBody := map[string]interface{}{
		"model":       "text-davinci-003", // You can use another model if you prefer
		"prompt":      prompt,
		"max_tokens":  1000,
		"temperature": 0.5,
	}

	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		http.Error(w, "Error preparing the request", http.StatusInternalServerError)
		return
	}

	// Make the POST request to OpenAI API
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/completions", bytes.NewBuffer(reqBodyJson))
	if err != nil {
		http.Error(w, "Error creating the request", http.StatusInternalServerError)
		return
	}

	// Add Authorization header with the OpenAI API key
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Get response from OpenAI
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error sending the request to OpenAI", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Decode the response from OpenAI
	var openAIResponse OpenAIResponse
	err = json.NewDecoder(resp.Body).Decode(&openAIResponse)
	if err != nil {
		http.Error(w, "Error decoding the response", http.StatusInternalServerError)
		return
	}

	// Extract the text from the OpenAI response (potential invariants)
	invariants := []string{}
	for _, choice := range openAIResponse.Choices {
		// Assuming the response is a plain text list of invariants
		invariants = append(invariants, strings.TrimSpace(choice.Text))
	}

	// Send the invariants back in the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"invariants": invariants,
	})
}
