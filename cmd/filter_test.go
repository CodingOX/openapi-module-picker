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

// buildFilterMockServer 创建一个包含两条路径的模拟 OpenAPI 服务器。
// 路径 /users 关联标签 "users"，路径 /orders 关联标签 "orders"。
func buildFilterMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
}

// TestRunFilter_Success 测试基本过滤功能：选中 "users" 标签后输出文件仅含 /users 路径。
func TestRunFilter_Success(t *testing.T) {
	mockServer := buildFilterMockServer()
	defer mockServer.Close()

	outputPath := filepath.Join(t.TempDir(), "output.json")

	// 捕获 stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := RunFilter([]string{"--url", mockServer.URL, "--tags", "users", "--output", outputPath})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	stdoutOutput := strings.TrimSpace(buf.String())

	if err != nil {
		t.Fatalf("RunFilter 失败: %v", err)
	}

	// 验证 stdout 输出为绝对路径
	absPath, _ := filepath.Abs(outputPath)
	if stdoutOutput != absPath {
		t.Errorf("stdout 应为绝对路径 %s，实际: %s", absPath, stdoutOutput)
	}

	// 验证输出文件内容
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("读取输出文件失败: %v", err)
	}

	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("解析输出 JSON 失败: %v", err)
	}

	paths, ok := doc["paths"].(map[string]interface{})
	if !ok {
		t.Fatal("输出文档应包含 paths 字段")
	}
	if _, ok := paths["/users"]; !ok {
		t.Error("输出应包含 /users 路径")
	}
	if _, ok := paths["/orders"]; ok {
		t.Error("输出不应包含 /orders 路径")
	}
}

// TestRunFilter_AutoCreateDir 测试输出路径含深层嵌套目录时自动创建父目录。
func TestRunFilter_AutoCreateDir(t *testing.T) {
	mockServer := buildFilterMockServer()
	defer mockServer.Close()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "deeply", "nested", "output.json")

	// 捕获 stdout，静默输出
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := RunFilter([]string{"--url", mockServer.URL, "--tags", "users", "--output", outputPath})

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("RunFilter 失败: %v", err)
	}

	// 验证文件存在于深层嵌套目录中
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("输出文件应存在: %s", outputPath)
	}
}

// TestRunFilter_MissingURL 测试缺少 --url 参数时应返回错误。
func TestRunFilter_MissingURL(t *testing.T) {
	err := RunFilter([]string{"--tags", "users", "--output", "out.json"})
	if err == nil {
		t.Fatal("缺少 --url 参数应返回错误")
	}
	if !strings.Contains(err.Error(), "--url") {
		t.Errorf("错误信息应提及 --url，实际: %v", err)
	}
}

// TestRunFilter_MissingTags 测试缺少 --tags 参数时应返回错误。
func TestRunFilter_MissingTags(t *testing.T) {
	err := RunFilter([]string{"--url", "http://example.com", "--output", "out.json"})
	if err == nil {
		t.Fatal("缺少 --tags 参数应返回错误")
	}
	if !strings.Contains(err.Error(), "--tags") {
		t.Errorf("错误信息应提及 --tags，实际: %v", err)
	}
}

// TestRunFilter_MissingOutput 测试缺少 --output 参数时应返回错误。
func TestRunFilter_MissingOutput(t *testing.T) {
	err := RunFilter([]string{"--url", "http://example.com", "--tags", "users"})
	if err == nil {
		t.Fatal("缺少 --output 参数应返回错误")
	}
	if !strings.Contains(err.Error(), "--output") {
		t.Errorf("错误信息应提及 --output，实际: %v", err)
	}
}

// TestRunFilter_MultipleTags 测试逗号分隔的多标签过滤，验证两条路径均保留。
func TestRunFilter_MultipleTags(t *testing.T) {
	mockServer := buildFilterMockServer()
	defer mockServer.Close()

	outputPath := filepath.Join(t.TempDir(), "output.json")

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := RunFilter([]string{"--url", mockServer.URL, "--tags", "users,orders", "--output", outputPath})

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("RunFilter 失败: %v", err)
	}

	data, _ := os.ReadFile(outputPath)
	var doc map[string]interface{}
	json.Unmarshal(data, &doc)

	paths := doc["paths"].(map[string]interface{})
	if _, ok := paths["/users"]; !ok {
		t.Error("输出应包含 /users 路径")
	}
	if _, ok := paths["/orders"]; !ok {
		t.Error("输出应包含 /orders 路径")
	}
}

// TestRunFilter_TagWithSpaces 测试标签含多余空格时 TrimSpace 正确裁剪。
func TestRunFilter_TagWithSpaces(t *testing.T) {
	mockServer := buildFilterMockServer()
	defer mockServer.Close()

	outputPath := filepath.Join(t.TempDir(), "output.json")

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := RunFilter([]string{"--url", mockServer.URL, "--tags", " users , orders ", "--output", outputPath})

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("RunFilter 失败: %v", err)
	}

	data, _ := os.ReadFile(outputPath)
	var doc map[string]interface{}
	json.Unmarshal(data, &doc)

	paths := doc["paths"].(map[string]interface{})
	if _, ok := paths["/users"]; !ok {
		t.Error("输出应包含 /users 路径（tag 空白裁剪）")
	}
	if _, ok := paths["/orders"]; !ok {
		t.Error("输出应包含 /orders 路径（tag 空白裁剪）")
	}
}

// TestRunFilter_EmptyTags 测试文档无标签时应返回 ErrNoTagsFound 哨兵错误。
func TestRunFilter_EmptyTags(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"openapi": "3.0.3",
			"paths":   map[string]interface{}{},
		})
	}))
	defer mockServer.Close()

	outputPath := filepath.Join(t.TempDir(), "output.json")

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := RunFilter([]string{"--url", mockServer.URL, "--tags", "users", "--output", outputPath})

	w.Close()
	os.Stdout = old

	if err == nil {
		t.Fatal("空标签应返回错误")
	}
	if !errors.Is(err, ErrNoTagsFound) {
		t.Errorf("期望 ErrNoTagsFound，实际: %v", err)
	}
}
