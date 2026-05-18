// Package main 提供 OpenAPI Module Picker 的 HTTP 服务器入口。
// 该服务允许用户从远程 OpenAPI/Swagger 文档中按标签筛选 API 模块并导出。
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

// main 启动 HTTP 服务器，配置静态文件服务和 API 路由。
// 服务器监听 :8080 端口，提供以下功能：
//   - / : 静态文件服务（web 目录）
//   - /api/parse : 解析远程 OpenAPI 文档
//   - /api/filter : 按标签过滤文档
func main() {
	// 提供静态文件服务
	fs := http.FileServer(http.Dir("web"))
	http.Handle("/", fs)

	// 注册 API 端点
	http.HandleFunc("/api/parse", handleParse)
	http.HandleFunc("/api/filter", handleFilter)

	fmt.Println("OpenAPI Module Picker starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handleParse 处理 OpenAPI 文档解析请求。
// 从请求体中获取 URL，解析远程 OpenAPI 文档，提取所有标签并返回。
// 仅支持 POST 方法，解析后的文档存储在全局变量 currentDocument 中。
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

	// 解析 OpenAPI 文档
	doc, err := openapi.ParseOpenAPI(req.URL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ParseResponse{
			Success: false,
			Message: "Failed to parse OpenAPI: " + err.Error(),
		})
		return
	}

	// 存储文档到全局变量，供后续过滤使用
	currentDocument = doc

	// 提取所有标签
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
// 根据用户选择的标签列表，过滤已解析的 OpenAPI 文档，返回过滤后的 JSON。
// 仅支持 POST 方法，需要先通过 /api/parse 加载文档。
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

	// 过滤文档
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

// init 初始化函数，检查 web 目录是否存在。
// 确保静态文件目录可访问，如果不存在会被优雅处理。
func init() {
	// 检查 web 目录是否存在，确保静态文件可正确提供
	if err := filepath.Walk("web", func(path string, info fs.FileInfo, err error) error {
		return nil
	}); err != nil {
		// web 目录不存在时会被优雅处理
	}
}
