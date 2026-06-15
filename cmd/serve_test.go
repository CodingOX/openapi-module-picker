package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"openapi-module-picker/openapi"
)

// ===== handleParse 测试 =====

// TestHandleParse_Success 测试成功解析 OpenAPI 文档。
func TestHandleParse_Success(t *testing.T) {
	// 启动 mock OpenAPI server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"openapi": "3.0.3",
			"info":    map[string]interface{}{"title": "Test", "version": "1.0"},
			"paths": map[string]interface{}{
				"/test": map[string]interface{}{
					"get": map[string]interface{}{
						"tags": []interface{}{"test"},
					},
				},
			},
		})
	}))
	defer mockServer.Close()

	// 重置全局状态
	currentDocument = nil

	body, _ := json.Marshal(ParseRequest{URL: mockServer.URL})
	req := httptest.NewRequest(http.MethodPost, "/api/parse", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleParse(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}

	var resp ParseResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Errorf("期望 success=true，实际消息: %s", resp.Message)
	}
	if resp.Version != "3.0" {
		t.Errorf("期望版本 3.0，实际 %s", resp.Version)
	}
	if len(resp.Tags) != 1 || resp.Tags[0] != "test" {
		t.Errorf("期望标签 [test]，实际 %v", resp.Tags)
	}
}

// TestHandleParse_MethodNotAllowed 测试非 POST 方法应返回 405。
func TestHandleParse_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/parse", nil)
	w := httptest.NewRecorder()

	handleParse(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("期望状态码 405，实际 %d", w.Code)
	}
}

// TestHandleParse_EmptyURL 测试空 URL 应返回 400。
func TestHandleParse_EmptyURL(t *testing.T) {
	body, _ := json.Marshal(ParseRequest{URL: ""})
	req := httptest.NewRequest(http.MethodPost, "/api/parse", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleParse(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	var resp ParseResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Success {
		t.Error("空 URL 应返回 success=false")
	}
}

// TestHandleParse_InvalidBody 测试无效请求体应返回 400。
func TestHandleParse_InvalidBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/parse", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleParse(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}
}

// TestHandleParse_ServerError 测试服务器错误时不应返回 200。
func TestHandleParse_ServerError(t *testing.T) {
	// 启动一个返回 500 的 mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	body, _ := json.Marshal(ParseRequest{URL: mockServer.URL})
	req := httptest.NewRequest(http.MethodPost, "/api/parse", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleParse(w, req)

	// ParseOpenAPI 会尝试读取 body，可能成功解析为无效 JSON，或连接失败
	// 这里主要验证不会 panic，状态码应为 400（解析失败）
	if w.Code == http.StatusOK {
		t.Error("服务器错误时不应返回 200")
	}
}

// ===== handleFilter 测试 =====

// TestHandleFilter_Success 测试成功过滤文档。
func TestHandleFilter_Success(t *testing.T) {
	// 预设全局文档
	currentDocument = &openapi.OpenAPIDocument{
		Version: "3.0",
		Raw: map[string]interface{}{
			"openapi": "3.0.3",
			"info":    map[string]interface{}{"title": "Test", "version": "1.0"},
			"paths": map[string]interface{}{
				"/users": map[string]interface{}{
					"get": map[string]interface{}{"tags": []interface{}{"users"}},
				},
				"/products": map[string]interface{}{
					"get": map[string]interface{}{"tags": []interface{}{"products"}},
				},
			},
		},
	}

	body, _ := json.Marshal(FilterRequest{SelectedTags: []string{"users"}})
	req := httptest.NewRequest(http.MethodPost, "/api/filter", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleFilter(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}

	var resp FilterResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Errorf("期望 success=true，实际消息: %s", resp.Message)
	}
	if resp.Data == "" {
		t.Error("过滤结果 data 不应为空")
	}
}

// TestHandleFilter_NoDocumentLoaded 测试未加载文档时应返回 400。
func TestHandleFilter_NoDocumentLoaded(t *testing.T) {
	currentDocument = nil

	body, _ := json.Marshal(FilterRequest{SelectedTags: []string{"users"}})
	req := httptest.NewRequest(http.MethodPost, "/api/filter", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleFilter(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}

	var resp FilterResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Success {
		t.Error("无文档时应返回 success=false")
	}
}

// TestHandleFilter_MethodNotAllowed 测试非 POST 方法应返回 405。
func TestHandleFilter_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/filter", nil)
	w := httptest.NewRecorder()

	handleFilter(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("期望状态码 405，实际 %d", w.Code)
	}
}

// TestHandleFilter_EmptyTags 测试空标签列表应返回 400。
func TestHandleFilter_EmptyTags(t *testing.T) {
	currentDocument = &openapi.OpenAPIDocument{
		Version: "3.0",
		Raw:     map[string]interface{}{"openapi": "3.0.3"},
	}

	body, _ := json.Marshal(FilterRequest{SelectedTags: []string{}})
	req := httptest.NewRequest(http.MethodPost, "/api/filter", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleFilter(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}
}

// TestHandleFilter_InvalidBody 测试无效请求体应返回 400。
func TestHandleFilter_InvalidBody(t *testing.T) {
	currentDocument = &openapi.OpenAPIDocument{
		Version: "3.0",
		Raw:     map[string]interface{}{"openapi": "3.0.3"},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/filter", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleFilter(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}
}
