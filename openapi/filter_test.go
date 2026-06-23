package openapi

import (
	"encoding/json"
	"testing"
)

// ===== 辅助函数 =====

// buildFilterTestDoc 构建测试用 OpenAPI 3.0 文档。
func buildFilterTestDoc() *OpenAPIDocument {
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
					"tags": []interface{}{"orders"},
				},
			},
		},
	}
	return &OpenAPIDocument{Version: "3.0", Raw: raw}
}

// ===== FilterByTags 测试 =====

// TestFilterByTags_SingleTag 测试单标签过滤。
func TestFilterByTags_SingleTag(t *testing.T) {
	doc := buildFilterTestDoc()

	result, err := doc.FilterByTags([]string{"users"})
	if err != nil {
		t.Fatalf("过滤失败: %v", err)
	}

	// 解析结果验证
	var filtered map[string]interface{}
	if err := json.Unmarshal(result, &filtered); err != nil {
		t.Fatalf("结果不是有效 JSON: %v", err)
	}

	paths, ok := filtered["paths"].(map[string]interface{})
	if !ok {
		t.Fatal("结果中缺少 paths 字段")
	}

	// 应保留 /users，移除 /products 和 /orders
	if _, exists := paths["/users"]; !exists {
		t.Error("/users 路径应被保留")
	}
	if _, exists := paths["/products"]; exists {
		t.Error("/products 路径应被移除")
	}
	if _, exists := paths["/orders"]; exists {
		t.Error("/orders 路径应被移除")
	}
}

// TestFilterByTags_MultipleTags 测试多标签过滤。
func TestFilterByTags_MultipleTags(t *testing.T) {
	doc := buildFilterTestDoc()

	result, err := doc.FilterByTags([]string{"users", "products"})
	if err != nil {
		t.Fatalf("过滤失败: %v", err)
	}

	var filtered map[string]interface{}
	json.Unmarshal(result, &filtered)
	paths := filtered["paths"].(map[string]interface{})

	// 应保留 /users 和 /products，移除 /orders
	if _, exists := paths["/users"]; !exists {
		t.Error("/users 路径应被保留")
	}
	if _, exists := paths["/products"]; !exists {
		t.Error("/products 路径应被保留")
	}
	if _, exists := paths["/orders"]; exists {
		t.Error("/orders 路径应被移除")
	}
}

// TestFilterByTags_NoMatchingTags 测试无匹配标签时应返回空路径。
func TestFilterByTags_NoMatchingTags(t *testing.T) {
	doc := buildFilterTestDoc()

	result, err := doc.FilterByTags([]string{"nonexistent"})
	if err != nil {
		t.Fatalf("过滤失败: %v", err)
	}

	var filtered map[string]interface{}
	json.Unmarshal(result, &filtered)
	paths := filtered["paths"].(map[string]interface{})

	if len(paths) != 0 {
		t.Errorf("无匹配标签时 paths 应为空，实际有 %d 个", len(paths))
	}
}

// TestFilterByTags_AllTags 测试选择所有标签时应保留所有路径。
func TestFilterByTags_AllTags(t *testing.T) {
	doc := buildFilterTestDoc()

	result, err := doc.FilterByTags([]string{"users", "products", "orders"})
	if err != nil {
		t.Fatalf("过滤失败: %v", err)
	}

	var filtered map[string]interface{}
	json.Unmarshal(result, &filtered)
	paths := filtered["paths"].(map[string]interface{})

	// 所有路径都应保留
	if len(paths) != 3 {
		t.Errorf("选择所有标签时应保留 3 个路径，实际有 %d 个", len(paths))
	}
}

// TestFilterByTags_PreservesInfo 测试过滤后应保留 info 字段。
func TestFilterByTags_PreservesInfo(t *testing.T) {
	doc := buildFilterTestDoc()

	result, err := doc.FilterByTags([]string{"users"})
	if err != nil {
		t.Fatalf("过滤失败: %v", err)
	}

	var filtered map[string]interface{}
	json.Unmarshal(result, &filtered)

	// info 字段应被保留
	info, exists := filtered["info"]
	if !exists {
		t.Error("info 字段应被保留")
	}
	infoMap, ok := info.(map[string]interface{})
	if !ok {
		t.Fatal("info 字段格式错误")
	}
	if infoMap["title"] != "Test API" {
		t.Error("info.title 应为 'Test API'")
	}
}

// TestFilterByTags_PreservesOpenAPIVersion 测试过滤后应保留 OpenAPI 版本字段。
func TestFilterByTags_PreservesOpenAPIVersion(t *testing.T) {
	doc := buildFilterTestDoc()

	result, err := doc.FilterByTags([]string{"users"})
	if err != nil {
		t.Fatalf("过滤失败: %v", err)
	}

	var filtered map[string]interface{}
	json.Unmarshal(result, &filtered)

	// openapi 版本字段应被保留
	version, exists := filtered["openapi"]
	if !exists {
		t.Error("openapi 字段应被保留")
	}
	if version != "3.0.3" {
		t.Errorf("openapi 版本应为 '3.0.3'，实际为 '%v'", version)
	}
}

// TestFilterByTags_Swagger20 测试 Swagger 2.0 文档的过滤。
func TestFilterByTags_Swagger20(t *testing.T) {
	raw := map[string]interface{}{
		"swagger": "2.0",
		"info":    map[string]interface{}{"title": "Swagger API", "version": "1.0"},
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

	result, err := doc.FilterByTags([]string{"pets"})
	if err != nil {
		t.Fatalf("过滤失败: %v", err)
	}

	var filtered map[string]interface{}
	json.Unmarshal(result, &filtered)
	paths := filtered["paths"].(map[string]interface{})

	if _, exists := paths["/pets"]; !exists {
		t.Error("/pets 路径应被保留")
	}
	if _, exists := paths["/store"]; exists {
		t.Error("/store 路径应被移除")
	}
}

// TestFilterByTags_MultiTagOperation 测试一个 operation 有多个 tag 的情况。
// 选择其中任意一个 tag 都应保留该路径。
func TestFilterByTags_MultiTagOperation(t *testing.T) {
	raw := map[string]interface{}{
		"openapi": "3.0.3",
		"paths": map[string]interface{}{
			"/mixed": map[string]interface{}{
				"get": map[string]interface{}{
					"tags": []interface{}{"tagA", "tagB"},
				},
			},
		},
	}
	doc := &OpenAPIDocument{Version: "3.0", Raw: raw}

	result, err := doc.FilterByTags([]string{"tagA"})
	if err != nil {
		t.Fatalf("过滤失败: %v", err)
	}

	var filtered map[string]interface{}
	json.Unmarshal(result, &filtered)
	paths := filtered["paths"].(map[string]interface{})

	if _, exists := paths["/mixed"]; !exists {
		t.Error("选择 tagA 时 /mixed 路径应被保留（operation 同时属于 tagA 和 tagB）")
	}
}

// TestFilterByTags_EmptyDocument 测试空文档过滤后仍应返回有效 JSON。
func TestFilterByTags_EmptyDocument(t *testing.T) {
	doc := &OpenAPIDocument{
		Version: "3.0",
		Raw:     map[string]interface{}{},
	}

	result, err := doc.FilterByTags([]string{"any"})
	if err != nil {
		t.Fatalf("过滤失败: %v", err)
	}

	// 空文档过滤后仍应返回有效 JSON
	var filtered map[string]interface{}
	if err := json.Unmarshal(result, &filtered); err != nil {
		t.Fatalf("结果不是有效 JSON: %v", err)
	}
}

// TestFilterByTags_PrunesUnreferencedSchemas 测试过滤后裁剪未引用的 schema。
// exam 和 user 两个 tag 各自使用不同的 schema，以及一个共享 schema。
// 按 exam 过滤后，UserResponse 应被移除，ExamResponse 和 SharedModel 应保留。
func TestFilterByTags_PrunesUnreferencedSchemas(t *testing.T) {
	raw := map[string]interface{}{
		"openapi": "3.0.3",
		"info":    map[string]interface{}{"title": "Test", "version": "1.0"},
		"paths": map[string]interface{}{
			"/exams": map[string]interface{}{
				"get": map[string]interface{}{
					"tags": []interface{}{"exam"},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ExamResponse",
									},
								},
							},
						},
					},
				},
			},
			"/users": map[string]interface{}{
				"get": map[string]interface{}{
					"tags": []interface{}{"user"},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/UserResponse",
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
				"ExamResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{"type": "integer"},
						"shared": map[string]interface{}{
							"$ref": "#/components/schemas/SharedModel",
						},
					},
				},
				"UserResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{"type": "string"},
					},
				},
				"SharedModel": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"createdAt": map[string]interface{}{"type": "string"},
					},
				},
			},
		},
	}
	doc := &OpenAPIDocument{Version: "3.0", Raw: raw}

	result, err := doc.FilterByTags([]string{"exam"})
	if err != nil {
		t.Fatalf("过滤失败: %v", err)
	}

	var filtered map[string]interface{}
	if err := json.Unmarshal(result, &filtered); err != nil {
		t.Fatalf("结果不是有效 JSON: %v", err)
	}

	// 验证 paths：仅保留 /exams
	paths, _ := filtered["paths"].(map[string]interface{})
	if _, exists := paths["/exams"]; !exists {
		t.Error("/exams 路径应被保留")
	}
	if _, exists := paths["/users"]; exists {
		t.Error("/users 路径应被移除")
	}

	// 验证 schemas：ExamResponse 和 SharedModel 保留，UserResponse 移除
	components, _ := filtered["components"].(map[string]interface{})
	schemas, _ := components["schemas"].(map[string]interface{})

	if _, exists := schemas["ExamResponse"]; !exists {
		t.Error("ExamResponse schema 应被保留（直接被 exam 引用）")
	}
	if _, exists := schemas["SharedModel"]; !exists {
		t.Error("SharedModel schema 应被保留（被 ExamResponse 传递引用）")
	}
	if _, exists := schemas["UserResponse"]; exists {
		t.Error("UserResponse schema 应被移除（未被 exam 引用）")
	}
}
func TestFilterByTags_ValidJSON(t *testing.T) {
	doc := buildFilterTestDoc()

	result, err := doc.FilterByTags([]string{"users"})
	if err != nil {
		t.Fatalf("过滤失败: %v", err)
	}

	// json.MarshalIndent 生成的 JSON 应包含换行和缩进
	jsonStr := string(result)
	if len(jsonStr) == 0 {
		t.Error("结果不应为空")
	}

	// 验证可以被正常解析
	var parsed map[string]interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("结果不是有效 JSON: %v", err)
	}
}
