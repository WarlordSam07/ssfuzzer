package handlers

// Common request/response types
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

type Response struct {
	Success    bool     `json:"success"`
	Invariants []string `json:"invariants,omitempty"`
	Error      string   `json:"error,omitempty"`
}

type SolidityCodeRequest struct {
	SolidityCode string `json:"solidityCode"`
}
