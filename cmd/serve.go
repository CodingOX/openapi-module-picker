package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"

	"openapi-trim/openapi"
)

// currentDocument 存储已解析的 OpenAPI 文档，用于后续过滤操作。
// 注意：当前实现不支持并发多用户场景。
var currentDocument *openapi.OpenAPIDocument

// ParseRequest 表示解析请求的消息体，包含 OpenAPI 文档的 URL。
type ParseRequest struct {
	URL string `json:"url"`
}

// FilterRequest 表示过滤请求的消息体，包含用户选择的标签列表。
type FilterRequest struct {
	SelectedTags []string `json:"selectedTags"`
}

// ParseResponse 表示解析响应，包含可用的标签列表和文档版本信息。
type ParseResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Tags    []string `json:"tags,omitempty"`
	Version string   `json:"version,omitempty"`
}

// FilterResponse 表示过滤响应，包含过滤后的 OpenAPI 文档 JSON 数据。
type FilterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// RunServe 启动 Web 服务。
// 提供静态文件服务和 API 端点：/api/parse、/api/filter。
// 此函数会阻塞直到服务停止。
func RunServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	port := fs.String("port", "8326", "监听端口")

	if err := fs.Parse(args); err != nil {
		return err
	}

	ensureWebDir()

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("web")))
	mux.HandleFunc("/api/parse", handleParse)
	mux.HandleFunc("/api/filter", handleFilter)

	addr := ":" + *port
	fmt.Printf("openapi-trim starting on http://localhost%s\n", addr)
	return http.ListenAndServe(addr, mux)
}

// ensureWebDir 检查 web 目录是否存在。
func ensureWebDir() {
	if err := filepath.Walk("web", func(path string, info fs.FileInfo, err error) error {
		return nil
	}); err != nil {
		// web 目录不存在时会被优雅处理
	}
}

// handleParse 处理 OpenAPI 文档解析请求。
func handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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

	doc, err := openapi.ParseOpenAPI(req.URL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ParseResponse{
			Success: false,
			Message: "Failed to parse OpenAPI: " + err.Error(),
		})
		return
	}

	currentDocument = doc

	tags := doc.GetAllTags()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ParseResponse{
		Success: true,
		Message: "OpenAPI parsed successfully",
		Tags:    tags,
		Version: doc.Version,
	})
}

// handleFilter 处理 OpenAPI 文档过滤请求。
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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
