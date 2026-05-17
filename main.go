package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"openapi-module-picker/openapi"
)

type parseRequest struct {
	URL string `json:"url"`
}

type filterRequest struct {
	SelectedTags []string `json:"selectedTags"`
}

type appState struct {
	mu       sync.RWMutex
	document map[string]any
}

func main() {
	state := &appState{}
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("web")))
	mux.HandleFunc("/api/parse", state.handleParse)
	mux.HandleFunc("/api/filter", state.handleFilter)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("OpenAPI Module Picker running on http://localhost:%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func (s *appState) handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "message": "method not allowed"})
		return
	}

	body, err := parseRequestBody(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": err.Error()})
		return
	}

	doc, err := openapi.ParseDocument(body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": err.Error()})
		return
	}

	tags := openapi.ExtractTags(doc)
	version := openapi.DetectVersion(doc)

	s.mu.Lock()
	s.document = doc
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "OpenAPI parsed successfully",
		"tags":    tags,
		"version": version,
	})
}

func parseRequestBody(r *http.Request) ([]byte, error) {
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(64 << 20); err != nil {
			return nil, fmt.Errorf("invalid multipart form: %w", err)
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			return nil, errors.New("missing file upload")
		}
		defer file.Close()
		body, err := io.ReadAll(file)
		if err != nil {
			return nil, errors.New("failed to read uploaded file")
		}
		return body, nil
	}

	var req parseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.New("invalid request body")
	}
	u, err := url.Parse(req.URL)
	if err != nil || u.Scheme == "" || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, errors.New("invalid url")
	}
	if !isSafeRemoteHost(u.Hostname()) {
		return nil, errors.New("url host is not allowed")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(req.URL) // #nosec G107 - Host restrictions above block localhost and private ranges.
	if err != nil {
		return nil, fmt.Errorf("failed to fetch url: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to fetch url: http %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("failed to read response body")
	}
	return body, nil
}

func isSafeRemoteHost(host string) bool {
	if host == "" {
		return false
	}

	if strings.EqualFold(host, "localhost") {
		return false
	}

	if ip := net.ParseIP(host); ip != nil {
		return !isPrivateOrLocalIP(ip)
	}

	return true
}

func isPrivateOrLocalIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsPrivate() || ip.IsUnspecified()
}

func (s *appState) handleFilter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "message": "method not allowed"})
		return
	}

	var req filterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "invalid request body"})
		return
	}

	if len(req.SelectedTags) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "at least one tag must be selected"})
		return
	}

	s.mu.RLock()
	doc := s.document
	s.mu.RUnlock()

	if doc == nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "message": "no OpenAPI document loaded"})
		return
	}

	filtered, err := openapi.FilterByTags(doc, req.SelectedTags)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": err.Error()})
		return
	}

	bytes, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "failed to encode filtered document"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "OpenAPI filtered successfully",
		"data":    string(bytes),
	})
}

func writeJSON(w http.ResponseWriter, status int, payload map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
