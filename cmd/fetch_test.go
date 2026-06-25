package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunFetch_Success 测试从 URL 成功获取标签列表的场景。
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

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := RunFetch([]string{"--url", mockServer.URL})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if err != nil {
		t.Fatalf("RunFetch 失败: %v", err)
	}
	if !strings.Contains(output, "orders") || !strings.Contains(output, "users") {
		t.Errorf("stdout 应包含 orders 和 users 标签，实际输出: %s", output)
	}
}

// TestRunFetch_FromFile 测试从本地文件成功提取标签。
func TestRunFetch_FromFile(t *testing.T) {
	doc := map[string]interface{}{
		"openapi": "3.0.3",
		"info":    map[string]interface{}{"title": "Test", "version": "1.0"},
		"paths": map[string]interface{}{
			"/items": map[string]interface{}{
				"get": map[string]interface{}{
					"tags": []interface{}{"items"},
				},
			},
		},
	}

	tmpFile := filepath.Join(t.TempDir(), "test.json")
	data, _ := json.Marshal(doc)
	os.WriteFile(tmpFile, data, 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := RunFetch([]string{"--file", tmpFile})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if err != nil {
		t.Fatalf("RunFetch --file 失败: %v", err)
	}
	if !strings.Contains(output, "items") {
		t.Errorf("stdout 应包含 items 标签，实际输出: %s", output)
	}
}

// TestRunFetch_MissingSource 测试缺少 --file 和 --url 时应返回错误。
func TestRunFetch_MissingSource(t *testing.T) {
	err := RunFetch([]string{})
	if err == nil {
		t.Fatal("缺少数据源应返回错误")
	}
	if !strings.Contains(err.Error(), "--file") && !strings.Contains(err.Error(), "--url") {
		t.Errorf("错误信息应提及 --file 或 --url，实际: %v", err)
	}
}

// TestRunFetch_BothFileAndURL 测试同时指定 --file 和 --url 时应返回错误。
func TestRunFetch_BothFileAndURL(t *testing.T) {
	err := RunFetch([]string{"--file", "/tmp/test.json", "--url", "http://example.com"})
	if err == nil {
		t.Fatal("同时指定 --file 和 --url 应返回错误")
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

// TestRunFetch_EmptyTags 测试文档无标签时应返回 ErrNoTagsFound。
func TestRunFetch_EmptyTags(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"openapi": "3.0.3",
			"paths":   map[string]interface{}{},
		})
	}))
	defer mockServer.Close()

	err := RunFetch([]string{"--url", mockServer.URL})
	if err == nil {
		t.Fatal("空标签应返回错误")
	}
	if !errors.Is(err, ErrNoTagsFound) {
		t.Errorf("期望 ErrNoTagsFound，实际: %v", err)
	}
}
