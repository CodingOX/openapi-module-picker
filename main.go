package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"openapi-module-picker/openapi"
	"path/filepath"
)

// Global variable to store parsed document
var currentDocument *openapi.OpenAPIDocument

// ParseRequest represents the request body for parsing
type ParseRequest struct {
	URL string `json:"url"`
}

// FilterRequest represents the request body for filtering
type FilterRequest struct {
	SelectedTags []string `json:"selectedTags"`
}

// ParseResponse represents the response with available tags
type ParseResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Tags    []string `json:"tags,omitempty"`
	Version string   `json:"version,omitempty"`
}

// FilterResponse represents the filtered OpenAPI document
type FilterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

func main() {
	// Serve static files
	fs := http.FileServer(http.Dir("web"))
	http.Handle("/", fs)

	// API endpoints
	http.HandleFunc("/api/parse", handleParse)
	http.HandleFunc("/api/filter", handleFilter)

	fmt.Println("OpenAPI Module Picker starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handleParse handles the parsing of OpenAPI from URL
func handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req ParseRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ParseResponse{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	if req.URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ParseResponse{
			Success: false,
			Message: "URL is required",
		})
		return
	}

	// Parse OpenAPI document
	doc, err := openapi.ParseOpenAPI(req.URL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ParseResponse{
			Success: false,
			Message: "Failed to parse OpenAPI: " + err.Error(),
		})
		return
	}

	// Store the document globally for filtering
	currentDocument = doc

	// Get all tags
	tags := doc.GetAllTags()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ParseResponse{
		Success: true,
		Message: "OpenAPI parsed successfully",
		Tags:    tags,
		Version: doc.Version,
	})
}

// handleFilter handles filtering and returning the filtered document
func handleFilter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if currentDocument == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(FilterResponse{
			Success: false,
			Message: "No OpenAPI document loaded. Please parse first.",
		})
		return
	}

	var req FilterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(FilterResponse{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	if len(req.SelectedTags) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(FilterResponse{
			Success: false,
			Message: "At least one tag must be selected",
		})
		return
	}

	// Filter the document
	filtered, err := currentDocument.FilterByTags(req.SelectedTags)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(FilterResponse{
			Success: false,
			Message: "Failed to filter OpenAPI: " + err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(FilterResponse{
		Success: true,
		Message: "OpenAPI filtered successfully",
		Data:    string(filtered),
	})
}

func init() {
	// Change to the directory where the binary is located
	// This ensures static files are found correctly
	if err := filepath.Walk("web", func(path string, info fs.FileInfo, err error) error {
		return nil
	}); err != nil {
		// If web directory doesn't exist, it will be created or handled gracefully
	}
}
