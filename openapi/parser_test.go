package openapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"
)

// ===== detectOpenAPIVersion 测试 =====

// TestDetectOpenAPIVersion_30 测试 OpenAPI 3.0 版本检测。
func TestDetectOpenAPIVersion_30(t *testing.T) {
	doc := map[string]interface{}{
		"openapi": "3.0.3",
	}
	version := detectOpenAPIVersion(doc)
	if version != "3.0" {
		t.Errorf("期望版本 3.0，实际得到 %s", version)
	}
}

// TestDetectOpenAPIVersion_31 测试 OpenAPI 3.1 版本检测。
func TestDetectOpenAPIVersion_31(t *testing.T) {
	doc := map[string]interface{}{
		"openapi": "3.1.0",
	}
	version := detectOpenAPIVersion(doc)
	if version != "3.0" {
		t.Errorf("期望版本 3.0，实际得到 %s", version)
	}
}

// TestDetectOpenAPIVersion_Swagger20 测试 Swagger 2.0 版本检测。
func TestDetectOpenAPIVersion_Swagger20(t *testing.T) {
	doc := map[string]interface{}{
		"swagger": "2.0",
	}
	version := detectOpenAPIVersion(doc)
	if version != "2.0" {
		t.Errorf("期望版本 2.0，实际得到 %s", version)
	}
}

// TestDetectOpenAPIVersion_DefaultTo20 测试默认版本应为 2.0。
func TestDetectOpenAPIVersion_DefaultTo20(t *testing.T) {
	doc := map[string]interface{}{
		"info": map[string]interface{}{"title": "Test"},
	}
	version := detectOpenAPIVersion(doc)
	if version != "2.0" {
		t.Errorf("期望默认版本 2.0，实际得到 %s", version)
	}
}

// TestDetectOpenAPIVersion_EmptyOpenAPI 测试空 openapi 字段应默认为 2.0。
func TestDetectOpenAPIVersion_EmptyOpenAPI(t *testing.T) {
	doc := map[string]interface{}{
		"openapi": "",
	}
	version := detectOpenAPIVersion(doc)
	if version != "2.0" {
		t.Errorf("空 openapi 字段应默认为 2.0，实际得到 %s", version)
	}
}

// ===== GetAllTags 测试 =====

// buildOpenAPI30Doc 构建测试用 OpenAPI 3.0 文档。
func buildOpenAPI30Doc() *OpenAPIDocument {
	raw := map[string]interface{}{
		"openapi": "3.0.3",
		"info":    map[string]interface{}{"title": "Test API", "version": "1.0"},
		"paths": map[string]interface{}{
			"/users": map[string]interface{}{
				"get": map[string]interface{}{
					"tags": []interface{}{"users"},
				},
				"post": map[string]interface{}{
					"tags": []interface{}{"users"},
				},
			},
			"/products": map[string]interface{}{
				"get": map[string]interface{}{
					"tags": []interface{}{"products"},
				},
			},
			"/orders": map[string]interface{}{
				"get": map[string]interface{}{
					"tags": []interface{}{"orders", "users"},
				},
			},
		},
	}
	return &OpenAPIDocument{Version: "3.0", Raw: raw}
}

// TestGetAllTags_OpenAPI30 测试从 OpenAPI 3.0 文档提取所有标签。
func TestGetAllTags_OpenAPI30(t *testing.T) {
	doc := buildOpenAPI30Doc()
	tags := doc.GetAllTags()

	expected := []string{"orders", "products", "users"}
	sort.Strings(tags)
	sort.Strings(expected)

	if !reflect.DeepEqual(tags, expected) {
		t.Errorf("期望标签 %v，实际得到 %v", expected, tags)
	}
}

// TestGetAllTags_NoPaths 测试无 paths 时应返回空切片。
func TestGetAllTags_NoPaths(t *testing.T) {
	doc := &OpenAPIDocument{
		Version: "3.0",
		Raw:     map[string]interface{}{},
	}
	tags := doc.GetAllTags()
	if len(tags) != 0 {
		t.Errorf("无 paths 时应返回空切片，实际得到 %v", tags)
	}
}

// TestGetAllTags_NoTags 测试无 tags 时应返回空切片。
func TestGetAllTags_NoTags(t *testing.T) {
	doc := &OpenAPIDocument{
		Version: "3.0",
		Raw: map[string]interface{}{
			"paths": map[string]interface{}{
				"/test": map[string]interface{}{
					"get": map[string]interface{}{},
				},
			},
		},
	}
	tags := doc.GetAllTags()
	if len(tags) != 0 {
		t.Errorf("无 tags 时应返回空切片，实际得到 %v", tags)
	}
}

// TestGetAllTags_Swagger20 测试从 Swagger 2.0 文档提取所有标签。
func TestGetAllTags_Swagger20(t *testing.T) {
	raw := map[string]interface{}{
		"swagger": "2.0",
		"info":    map[string]interface{}{"title": "Test API", "version": "1.0"},
		"paths": map[string]interface{}{
			"/pets": map[string]interface{}{
				"get": map[string]interface{}{
					"tags": []interface{}{"pets"},
				},
			},
			"/store": map[string]interface{}{
				"get": map[string]interface{}{
					"tags": []interface{}{"store"},
				},
			},
		},
	}
	doc := &OpenAPIDocument{Version: "2.0", Raw: raw}
	tags := doc.GetAllTags()

	expected := []string{"pets", "store"}
	sort.Strings(tags)

	if !reflect.DeepEqual(tags, expected) {
		t.Errorf("期望标签 %v，实际得到 %v", expected, tags)
	}
}

// TestGetAllTags_Deduplication 测试标签去重功能。
// 同一个 tag 出现在多个路径和方法中，应去重。
func TestGetAllTags_Deduplication(t *testing.T) {
	raw := map[string]interface{}{
		"openapi": "3.0.3",
		"paths": map[string]interface{}{
			"/a": map[string]interface{}{
				"get":  map[string]interface{}{"tags": []interface{}{"shared"}},
				"post": map[string]interface{}{"tags": []interface{}{"shared"}},
			},
			"/b": map[string]interface{}{
				"get": map[string]interface{}{"tags": []interface{}{"shared"}},
			},
		},
	}
	doc := &OpenAPIDocument{Version: "3.0", Raw: raw}
	tags := doc.GetAllTags()

	if len(tags) != 1 || tags[0] != "shared" {
		t.Errorf("应去重为 [shared]，实际得到 %v", tags)
	}
}

// ===== ParseOpenAPI 测试 =====

// TestParseOpenAPI_Success 测试成功解析 OpenAPI 文档。
func TestParseOpenAPI_Success(t *testing.T) {
	// 启动 mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"openapi": "3.0.3",
			"info":    map[string]interface{}{"title": "Mock API", "version": "1.0"},
			"paths":   map[string]interface{}{},
		})
	}))
	defer server.Close()

	doc, err := ParseOpenAPI(server.URL)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	if doc.Version != "3.0" {
		t.Errorf("期望版本 3.0，实际得到 %s", doc.Version)
	}
}

// TestParseOpenAPI_InvalidURL 测试无效 URL 应返回错误。
func TestParseOpenAPI_InvalidURL(t *testing.T) {
	_, err := ParseOpenAPI("http://localhost:99999/nonexistent")
	if err == nil {
		t.Error("无效 URL 应返回错误")
	}
}

// TestParseOpenAPI_InvalidJSON 测试无效 JSON 应返回错误。
func TestParseOpenAPI_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	_, err := ParseOpenAPI(server.URL)
	if err == nil {
		t.Error("无效 JSON 应返回错误")
	}
}

// ===== tagsFromMap 测试 =====

// TestTagsFromMap 测试标签映射转换为排序后的切片。
func TestTagsFromMap(t *testing.T) {
	m := map[string]bool{
		"beta":  true,
		"alpha": true,
		"gamma": true,
	}
	result := tagsFromMap(m)

	expected := []string{"alpha", "beta", "gamma"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("期望 %v，实际得到 %v", expected, result)
	}
}

// TestTagsFromMap_Empty 测试空映射应返回空切片。
func TestTagsFromMap_Empty(t *testing.T) {
	m := map[string]bool{}
	result := tagsFromMap(m)
	if len(result) != 0 {
		t.Errorf("空 map 应返回空切片，实际得到 %v", result)
	}
}
