package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestRunFetch_Success 测试成功获取标签列表的场景。
func TestRunFetch_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"openapi": "3.0.3",
			"info":    map[string]interface{}{"title": "Test", "version": "1.0"},
			"paths": map[string]interface{}{
				"/users": map[string]interface{}{
					"get": map[string]interface{}{
						"tags": []interface{}{"users"},
					},
				},
				"/orders": map[string]interface{}{
					"get": map[string]interface{}{
						"tags": []interface{}{"orders"},
					},
				},
			},
		})
	}))
	defer mockServer.Close()

	err := RunFetch([]string{"--url", mockServer.URL})
	if err != nil {
		t.Fatalf("RunFetch 失败: %v", err)
	}
}

// TestRunFetch_MissingURL 测试缺少 --url 参数时应返回错误。
func TestRunFetch_MissingURL(t *testing.T) {
	err := RunFetch([]string{})
	if err == nil {
		t.Fatal("缺少 --url 参数应返回错误")
	}
	if !strings.Contains(err.Error(), "--url") {
		t.Errorf("错误信息应提及 --url，实际: %v", err)
	}
}

// TestRunFetch_InvalidURL 测试无效 URL 时应返回错误。
func TestRunFetch_InvalidURL(t *testing.T) {
	err := RunFetch([]string{"--url", "http://localhost:99999/nonexistent"})
	if err == nil {
		t.Fatal("无效 URL 应返回错误")
	}
}

// TestRunFetch_InvalidFlag 测试无效 flag 时应返回错误。
func TestRunFetch_InvalidFlag(t *testing.T) {
	err := RunFetch([]string{"--unknown", "value"})
	if err == nil {
		t.Fatal("无效 flag 应返回错误")
	}
}
