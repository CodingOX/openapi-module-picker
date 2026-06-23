package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// buildSummaryMockServer 创建一个包含多标签的模拟 OpenAPI 服务器。
func buildSummaryMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"openapi": "3.0.3",
			"info":    map[string]interface{}{"title": "测试 API", "version": "1.0"},
			"tags": []interface{}{
				map[string]interface{}{"name": "exam", "description": "考试管理"},
				map[string]interface{}{"name": "user", "description": "用户管理"},
			},
			"paths": map[string]interface{}{
				"/exams": map[string]interface{}{
					"get": map[string]interface{}{
						"tags":    []interface{}{"exam"},
						"summary": "查询考试列表",
						"responses": map[string]interface{}{
							"200": map[string]interface{}{
								"description": "OK",
								"content": map[string]interface{}{
									"application/json": map[string]interface{}{
										"schema": map[string]interface{}{
											"$ref": "#/components/schemas/ExamPageResult",
										},
									},
								},
							},
						},
					},
				},
				"/users": map[string]interface{}{
					"get": map[string]interface{}{
						"tags":    []interface{}{"user"},
						"summary": "查询用户列表",
						"responses": map[string]interface{}{
							"200": map[string]interface{}{
								"description": "OK",
								"content": map[string]interface{}{
									"application/json": map[string]interface{}{
										"schema": map[string]interface{}{
											"$ref": "#/components/schemas/UserPageResult",
										},
									},
								},
							},
						},
					},
				},
			},
			"components": map[string]interface{}{
				"schemas": map[string]interface{}{
					"ExamPageResult": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"total": map[string]interface{}{"type": "integer", "title": "总数"},
							"items": map[string]interface{}{
								"type":  "array",
								"items": map[string]interface{}{"$ref": "#/components/schemas/ExamItem"},
							},
						},
					},
					"ExamItem": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"id":    map[string]interface{}{"type": "integer"},
							"title": map[string]interface{}{"type": "string"},
						},
					},
					"UserPageResult": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"total": map[string]interface{}{"type": "integer"},
							"users": map[string]interface{}{
								"type":  "array",
								"items": map[string]interface{}{"$ref": "#/components/schemas/UserInfo"},
							},
						},
					},
					"UserInfo": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"id":   map[string]interface{}{"type": "integer"},
							"name": map[string]interface{}{"type": "string"},
						},
					},
				},
			},
		})
	}))
}

// writeTestJSON 将 map 写入临时 JSON 文件，返回路径。
func writeTestJSON(t *testing.T, data map[string]interface{}) string {
	t.Helper()
	path := t.TempDir() + "/test.json"
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(data); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}
	return path
}

// captureStdout 捕获函数执行期间的 stdout 输出。
func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return strings.TrimSpace(buf.String())
}

// ===== RunSummary 测试 =====

func TestRunSummary_FromURL(t *testing.T) {
	mockServer := buildSummaryMockServer()
	defer mockServer.Close()

	output := captureStdout(func() {
		if err := RunSummary([]string{"--url", mockServer.URL}); err != nil {
			t.Fatalf("RunSummary 失败: %v", err)
		}
	})

	if !strings.Contains(output, "测试 API") {
		t.Error("输出应包含 API 标题")
	}
	if !strings.Contains(output, "exam") {
		t.Error("输出应包含 exam 标签")
	}
	if !strings.Contains(output, "user") {
		t.Error("输出应包含 user 标签")
	}
}

func TestRunSummary_FromFile(t *testing.T) {
	mockServer := buildSummaryMockServer()
	defer mockServer.Close()
	// 从 URL 得到 JSON 写入临时文件
	filePath := writeTestJSON(t, map[string]interface{}{
		"openapi": "3.0.3",
		"info":    map[string]interface{}{"title": "文件 API", "version": "2.0"},
		"paths": map[string]interface{}{
			"/test": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":    []interface{}{"test-tag"},
					"summary": "测试接口",
				},
			},
		},
		"components": map[string]interface{}{
			"schemas": map[string]interface{}{},
		},
	})

	output := captureStdout(func() {
		if err := RunSummary([]string{"--file", filePath}); err != nil {
			t.Fatalf("RunSummary 失败: %v", err)
		}
	})

	if !strings.Contains(output, "文件 API") {
		t.Error("输出应包含文件中的 API 标题")
	}
	if !strings.Contains(output, "test-tag") {
		t.Error("输出应包含 test-tag")
	}
}

func TestRunSummary_MissingSource(t *testing.T) {
	err := RunSummary([]string{})
	if err == nil {
		t.Fatal("缺少参数应返回错误")
	}
	if !strings.Contains(err.Error(), "--file") && !strings.Contains(err.Error(), "--url") {
		t.Errorf("错误信息应提及 --file 或 --url，实际: %v", err)
	}
}

// ===== RunList 测试 =====

func TestRunList_WithTags(t *testing.T) {
	mockServer := buildSummaryMockServer()
	defer mockServer.Close()

	output := captureStdout(func() {
		if err := RunList([]string{"--url", mockServer.URL, "--tags", "exam"}); err != nil {
			t.Fatalf("RunList 失败: %v", err)
		}
	})

	if !strings.Contains(output, "GET") {
		t.Error("输出应包含 HTTP 方法")
	}
	if !strings.Contains(output, "/exams") {
		t.Error("输出应包含 /exams 路径")
	}
	if strings.Contains(output, "/users") {
		t.Error("仅筛选 exam 时不应包含 /users")
	}
}

func TestRunList_AllTags(t *testing.T) {
	mockServer := buildSummaryMockServer()
	defer mockServer.Close()

	output := captureStdout(func() {
		if err := RunList([]string{"--url", mockServer.URL}); err != nil {
			t.Fatalf("RunList 失败: %v", err)
		}
	})

	if !strings.Contains(output, "exam") && !strings.Contains(output, "user") {
		t.Error("列出全部标签时应包含 exam 和 user")
	}
}

func TestRunList_FromFile(t *testing.T) {
	filePath := writeTestJSON(t, map[string]interface{}{
		"openapi": "3.0.3",
		"paths": map[string]interface{}{
			"/ping": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":    []interface{}{"system"},
					"summary": "健康检查",
				},
			},
		},
		"components": map[string]interface{}{
			"schemas": map[string]interface{}{},
		},
	})

	output := captureStdout(func() {
		if err := RunList([]string{"--file", filePath, "--tags", "system"}); err != nil {
			t.Fatalf("RunList 失败: %v", err)
		}
	})

	if !strings.Contains(output, "/ping") {
		t.Error("输出应包含 /ping")
	}
}

// ===== RunDescribe 测试 =====

func TestRunDescribe_FromURL(t *testing.T) {
	mockServer := buildSummaryMockServer()
	defer mockServer.Close()

	output := captureStdout(func() {
		if err := RunDescribe([]string{"--url", mockServer.URL, "--path", "/exams", "--method", "get"}); err != nil {
			t.Fatalf("RunDescribe 失败: %v", err)
		}
	})

	if !strings.Contains(output, "GET /exams") {
		t.Error("输出应以 GET /exams 为标题")
	}
	if !strings.Contains(output, "总数") {
		t.Error("输出应包含展开的字段")
	}
}

func TestRunDescribe_FromFile(t *testing.T) {
	filePath := writeTestJSON(t, map[string]interface{}{
		"openapi": "3.0.3",
		"paths": map[string]interface{}{
			"/hello": map[string]interface{}{
				"post": map[string]interface{}{
					"summary": "打招呼",
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"name": map[string]interface{}{
											"type":  "string",
											"title": "姓名",
										},
									},
									"required": []interface{}{"name"},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "OK",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"message": map[string]interface{}{
												"type":  "string",
												"title": "问候语",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"components": map[string]interface{}{
			"schemas": map[string]interface{}{},
		},
	})

	output := captureStdout(func() {
		if err := RunDescribe([]string{"--file", filePath, "--path", "/hello", "--method", "post"}); err != nil {
			t.Fatalf("RunDescribe 失败: %v", err)
		}
	})

	if !strings.Contains(output, "POST /hello") {
		t.Error("输出应以 POST /hello 为标题")
	}
	if !strings.Contains(output, "姓名") {
		t.Error("输出应包含请求体字段说明")
	}
	if !strings.Contains(output, "问候语") {
		t.Error("输出应包含响应字段说明")
	}
}

func TestRunDescribe_MissingPath(t *testing.T) {
	err := RunDescribe([]string{"--url", "http://example.com"})
	if err == nil {
		t.Fatal("缺少 --path 应返回错误")
	}
	if !strings.Contains(err.Error(), "--path") {
		t.Errorf("错误信息应提及 --path，实际: %v", err)
	}
}

func TestRunDescribe_NotFound(t *testing.T) {
	mockServer := buildSummaryMockServer()
	defer mockServer.Close()

	output := captureStdout(func() {
		if err := RunDescribe([]string{"--url", mockServer.URL, "--path", "/nonexistent"}); err != nil {
			t.Fatalf("RunDescribe 失败: %v", err)
		}
	})

	if !strings.Contains(output, "错误") {
		t.Error("不存在的路径应返回错误提示")
	}
}

// ===== parseDoc 测试 =====

func TestParseDoc_BothFileAndURL(t *testing.T) {
	_, err := parseDoc("/tmp/fake.json", "http://example.com")
	if err == nil {
		t.Error("同时指定 --file 和 --url 应返回错误")
	}
}

func TestParseDoc_Neither(t *testing.T) {
	_, err := parseDoc("", "")
	if err == nil {
		t.Error("不指定 --file 和 --url 应返回错误")
	}
}
