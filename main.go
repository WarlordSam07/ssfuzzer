package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"solidity-invariant-fuzzer/handlers"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Now you can access the API key from the environment variables
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OpenAI API key not set in .env")
	}

	fmt.Println("API Key loaded successfully")

	r := mux.NewRouter()

	// Serve static files (JS, CSS, etc.)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Routes for the app
	r.HandleFunc("/", handlers.HomePage).Methods("GET")
	r.HandleFunc("/submit-code", handlers.SubmitCode).Methods("POST")
	r.HandleFunc("/generate-tests", handlers.GenerateTests).Methods("POST")
	r.HandleFunc("/run-echidna", handlers.RunEchidna).Methods("POST")

	// Start the server
	fmt.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
