package handlers

// Common request/response types
type TestGenerationRequest struct {
	SolidityCode string   `json:"solidityCode"`
	Invariants   []string `json:"invariants"`
}

// In TestGenerationResponse struct (types.go)
type TestGenerationResponse struct {
	Success      bool     `json:"success"`
	TestCode     string   `json:"testCode,omitempty"`
	Error        string   `json:"error,omitempty"`
	EchidnaFile  string   `json:"echidnaFile,omitempty"`
	TestFiles    []string `json:"testFiles,omitempty"`
	TestContents []string `json:"testContents,omitempty"`
	MarkdownDocs []string `json:"markdownDocs,omitempty"` // Add this field
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

// Add these to the existing types.go file

type EchidnaConfig struct {
	TestLimit       int      `json:"testLimit"`
	Coverage        bool     `json:"coverage"`
	CorpusDir       string   `json:"corpusDir"`
	TestMode        string   `json:"testMode"`
	CryticArgs      []string `json:"cryticArgs"`
	FilterFunctions []string `json:"filterFunctions"`
	SeqLen          int      `json:"seqLen"`
	ShrinkLimit     int      `json:"shrinkLimit"`
	Timeout         int      `json:"timeout"`
}

// Update EchidnaResponse to include more details
type EchidnaResponse struct {
	Success     bool   `json:"success"`
	Output      string `json:"output,omitempty"`
	Error       string `json:"error,omitempty"`
	Coverage    string `json:"coverage,omitempty"`
	TestsPassed int    `json:"testsPassed,omitempty"`
	TestsFailed int    `json:"testsFailed,omitempty"`
}

type EchidnaRequest struct {
	TestCode string `json:"testCode"`
}
