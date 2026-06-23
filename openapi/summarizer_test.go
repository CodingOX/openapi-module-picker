package openapi

import (
	"strings"
	"testing"
)

// ===== 辅助函数 =====

// buildSummaryTestDoc 构建测试用 OpenAPI 3.0 文档，覆盖多种场景。
func buildSummaryTestDoc() *OpenAPIDocument {
	raw := map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]interface{}{
			"title":   "测试 API",
			"version": "2.1.0",
		},
		"tags": []interface{}{
			map[string]interface{}{
				"name":        "exam",
				"description": "考试管理",
			},
			map[string]interface{}{
				"name":        "user",
				"description": "用户管理",
			},
		},
		"paths": map[string]interface{}{
			"/exams": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":    []interface{}{"exam"},
					"summary": "分页查询考试列表",
					"parameters": []interface{}{
						map[string]interface{}{
							"name":     "page",
							"in":       "query",
							"required": false,
							"schema": map[string]interface{}{
								"type":    "integer",
								"format":  "int32",
								"minimum": 1,
							},
						},
						map[string]interface{}{
							"name":     "pageSize",
							"in":       "query",
							"schema": map[string]interface{}{
								"type": "integer",
							},
						},
					},
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
				"post": map[string]interface{}{
					"tags":    []interface{}{"exam"},
					"summary": "创建考试",
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/CreateExamRequest",
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
										"$ref": "#/components/schemas/ExamDetail",
									},
								},
							},
						},
						"400": map[string]interface{}{
							"description": "参数错误",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ApiError",
									},
								},
							},
						},
					},
				},
			},
			"/exams/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":    []interface{}{"exam"},
					"summary": "查询考试详情",
					"parameters": []interface{}{
						map[string]interface{}{
							"name":     "id",
							"in":       "path",
							"required": true,
							"schema": map[string]interface{}{
								"type":   "integer",
								"format": "int64",
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "OK",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ExamDetail",
									},
								},
							},
						},
					},
				},
				"delete": map[string]interface{}{
					"tags":    []interface{}{"exam"},
					"summary": "删除考试",
					"responses": map[string]interface{}{
						"204": map[string]interface{}{
							"description": "删除成功",
						},
					},
				},
			},
			"/users": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":    []interface{}{"user"},
					"summary": "获取用户列表",
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
			"/mixed": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":    []interface{}{"exam", "user"},
					"summary": "混合接口（多标签）",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "OK",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ExamDetail",
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
						"total": map[string]interface{}{
							"type":    "integer",
							"format":  "int64",
							"title":   "总数",
						},
						"items": map[string]interface{}{
							"type":  "array",
							"title": "考试列表",
							"items": map[string]interface{}{
								"$ref": "#/components/schemas/ExamDetail",
							},
						},
					},
					"required": []interface{}{"total", "items"},
				},
				"CreateExamRequest": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"title": map[string]interface{}{
							"type":        "string",
							"title":       "考试名称",
							"description": "考试的名称，1-50 个字符",
							"maxLength":   50,
							"minLength":   1,
						},
						"duration": map[string]interface{}{
							"type":   "integer",
							"format": "int32",
							"title":  "考试时长（分钟）",
						},
						"teacher": map[string]interface{}{
							"$ref": "#/components/schemas/TeacherInfo",
						},
					},
					"required": []interface{}{"title", "duration"},
				},
				"ExamDetail": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":   "integer",
							"format": "int64",
							"title":  "考试 ID",
						},
						"title": map[string]interface{}{
							"type":  "string",
							"title": "考试名称",
						},
						"status": map[string]interface{}{
							"type":   "integer",
							"format": "int32",
							"title":  "状态",
						},
					},
				},
				"TeacherInfo": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":   "integer",
							"format": "int64",
							"title":  "教师 ID",
						},
						"name": map[string]interface{}{
							"type":  "string",
							"title": "教师姓名",
						},
					},
				},
				"UserPageResult": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"total": map[string]interface{}{
							"type":   "integer",
							"format": "int64",
							"title":  "用户总数",
						},
						"users": map[string]interface{}{
							"type":  "array",
							"title": "用户列表",
							"items": map[string]interface{}{
								"$ref": "#/components/schemas/UserInfo",
							},
						},
					},
				},
				"UserInfo": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":   "integer",
							"format": "int64",
						},
						"name": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"ApiError": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"code": map[string]interface{}{
							"type":   "integer",
							"format": "int32",
							"title":  "错误码",
						},
						"message": map[string]interface{}{
							"type":  "string",
							"title": "错误信息",
						},
					},
				},
			},
		},
	}
	return &OpenAPIDocument{Version: "3.0", Raw: raw}
}

// ===== GenerateSummary 测试 =====

func TestGenerateSummary_Basic(t *testing.T) {
	doc := buildSummaryTestDoc()
	s := NewSummarizer(doc)

	output := s.GenerateSummary()

	// 检查基本信息
	if !strings.Contains(output, "测试 API") {
		t.Error("输出应包含 API 标题")
	}
	if !strings.Contains(output, "2.1.0") {
		t.Error("输出应包含 API 版本号")
	}

	// 检查业务模块表格
	if !strings.Contains(output, "| exam") {
		t.Error("输出应包含 exam 标签")
	}
	if !strings.Contains(output, "| user") {
		t.Error("输出应包含 user 标签")
	}
	if !strings.Contains(output, "考试管理") {
		t.Error("输出应包含 exam 标签描述")
	}
	if !strings.Contains(output, "用户管理") {
		t.Error("输出应包含 user 标签描述")
	}

	// 检查数据模型
	if !strings.Contains(output, "ExamDetail") {
		t.Error("Top 数据模型应包含 ExamDetail（被引用 3 次）")
	}
	if !strings.Contains(output, "ApiError") {
		t.Error("Top 数据模型应包含 ApiError")
	}
}

func TestGenerateSummary_EmptyDoc(t *testing.T) {
	doc := &OpenAPIDocument{
		Version: "3.0",
		Raw:     map[string]interface{}{},
	}
	s := NewSummarizer(doc)
	output := s.GenerateSummary()

	if !strings.Contains(output, "v") {
		t.Error("空文档也应输出有效内容")
	}
}

// ===== ListByTags 测试 =====

func TestListByTags_SpecificTags(t *testing.T) {
	doc := buildSummaryTestDoc()
	s := NewSummarizer(doc)

	output := s.ListByTags([]string{"exam"})

	if !strings.Contains(output, "GET") {
		t.Error("输出应包含 HTTP 方法")
	}
	if !strings.Contains(output, "`/exams`") {
		t.Error("输出应包含 /exams 路径")
	}
	if !strings.Contains(output, "分页查询考试列表") {
		t.Error("输出应包含接口摘要")
	}
	if !strings.Contains(output, "`ExamDetail`") {
		t.Error("输出应包含涉及的数据模型 ExamDetail")
	}
	if strings.Contains(output, "获取用户列表") {
		t.Error("仅筛选 exam 时不应包含 user 接口")
	}
}

func TestListByTags_AllTags(t *testing.T) {
	doc := buildSummaryTestDoc()
	s := NewSummarizer(doc)

	output := s.ListByTags(nil)

	// 应包含所有标签和接口
	if !strings.Contains(output, "exam") {
		t.Error("输出应包含 exam 标签")
	}
	if !strings.Contains(output, "user") {
		t.Error("输出应包含 user 标签")
	}
	if !strings.Contains(output, "`/users`") {
		t.Error("输出应包含 /users 路径")
	}
}

func TestListByTags_EmptyDoc(t *testing.T) {
	doc := &OpenAPIDocument{
		Version: "3.0",
		Raw:     map[string]interface{}{},
	}
	s := NewSummarizer(doc)
	output := s.ListByTags([]string{"nonexistent"})

	if !strings.Contains(output, "未找到") && output != "" {
		t.Log("空文档和标签应给出友好提示")
	}
}

// ===== DescribeEndpoint 测试 =====

func TestDescribeEndpoint_GetWithParams(t *testing.T) {
	doc := buildSummaryTestDoc()
	s := NewSummarizer(doc)

	output := s.DescribeEndpoint("/exams", "get")

	// 标题
	if !strings.Contains(output, "GET /exams") {
		t.Error("输出应以方法+路径为标题")
	}
	if !strings.Contains(output, "分页查询考试列表") {
		t.Error("输出应包含接口摘要")
	}

	// 参数
	if !strings.Contains(output, "Query 参数") {
		t.Error("输出应包含 Query 参数段落")
	}
	if !strings.Contains(output, "`page`") {
		t.Error("输出应包含 page 参数")
	}
	if !strings.Contains(output, "`pageSize`") {
		t.Error("输出应包含 pageSize 参数")
	}

	// 响应
	if !strings.Contains(output, "200") {
		t.Error("输出应包含 200 状态码")
	}
	if !strings.Contains(output, "总数") {
		t.Error("输出应包含展开后的字段")
	}
}

func TestDescribeEndpoint_PostWithBody(t *testing.T) {
	doc := buildSummaryTestDoc()
	s := NewSummarizer(doc)

	output := s.DescribeEndpoint("/exams", "post")

	// 请求体
	if !strings.Contains(output, "Body") {
		t.Error("输出应包含 Body 段落")
	}
	if !strings.Contains(output, "必填") {
		t.Error("输出应标注请求体为必填")
	}
	if !strings.Contains(output, "`teacher.id`") {
		t.Error("输出应展开嵌套 $ref 为 teacher.id 字段")
	}
	if !strings.Contains(output, "教师 ID") {
		t.Error("输出应包含展开后字段的中文标题")
	}

	// 多个响应
	if !strings.Contains(output, "200") && !strings.Contains(output, "400") {
		t.Error("输出应包含多个响应状态码")
	}
}

func TestDescribeEndpoint_PathParam(t *testing.T) {
	doc := buildSummaryTestDoc()
	s := NewSummarizer(doc)

	output := s.DescribeEndpoint("/exams/{id}", "get")

	if !strings.Contains(output, "Path 参数") {
		t.Error("输出应包含 Path 参数段落")
	}
	if !strings.Contains(output, "`id`") {
		t.Error("输出应包含 id 路径参数")
	}
}

func TestDescribeEndpoint_NoContentResponse(t *testing.T) {
	doc := buildSummaryTestDoc()
	s := NewSummarizer(doc)

	output := s.DescribeEndpoint("/exams/{id}", "delete")

	if !strings.Contains(output, "204") {
		t.Error("输出应包含 204 状态码")
	}
	if !strings.Contains(output, "删除成功") {
		t.Error("输出应包含 204 的描述")
	}
}

func TestDescribeEndpoint_NotFound(t *testing.T) {
	doc := buildSummaryTestDoc()
	s := NewSummarizer(doc)

	output := s.DescribeEndpoint("/nonexistent", "get")

	if !strings.Contains(output, "错误") {
		t.Error("不存在的路径应返回错误提示")
	}
}

// ===== $ref 解析测试 =====

func TestResolveRef_Success(t *testing.T) {
	doc := buildSummaryTestDoc()
	s := NewSummarizer(doc)

	schema, err := s.resolveRef("#/components/schemas/ExamDetail")
	if err != nil {
		t.Fatalf("解析 $ref 失败: %v", err)
	}
	if schema["type"] != "object" {
		t.Errorf("期望 type=object，实际得到 %v", schema["type"])
	}
	props, ok := schema["properties"].(map[string]interface{})
	if !ok || len(props) == 0 {
		t.Error("ExamDetail 应有 properties")
	}
}

func TestResolveRef_NotFound(t *testing.T) {
	doc := buildSummaryTestDoc()
	s := NewSummarizer(doc)

	_, err := s.resolveRef("#/components/schemas/NonExistent")
	if err == nil {
		t.Error("不存在的 $ref 应返回错误")
	}
}

func TestResolveRef_InvalidRef(t *testing.T) {
	doc := buildSummaryTestDoc()
	s := NewSummarizer(doc)

	_, err := s.resolveRef("http://external/ref")
	if err == nil {
		t.Error("外部 $ref 应返回错误")
	}
}

// ===== flattenSchema 测试 =====

func TestFlattenSchema_BasicObject(t *testing.T) {
	doc := buildSummaryTestDoc()
	s := NewSummarizer(doc)

	schema, _ := s.resolveRef("#/components/schemas/ExamDetail")
	fields := s.flattenSchema(schema, nil)

	if len(fields) != 3 {
		t.Fatalf("ExamDetail 应有 3 个字段，实际 %d", len(fields))
	}

	// 检查是否包含 id、title、status
	fieldNames := make(map[string]bool)
	for _, f := range fields {
		fieldNames[f.Name] = true
	}
	for _, name := range []string{"id", "title", "status"} {
		if !fieldNames[name] {
			t.Errorf("缺少字段: %s", name)
		}
	}
}

func TestFlattenSchema_WithRefExpand(t *testing.T) {
	doc := buildSummaryTestDoc()
	s := NewSummarizer(doc)

	schema, _ := s.resolveRef("#/components/schemas/CreateExamRequest")
	fields := s.flattenSchema(schema, nil)

	hasTeacherID := false
	hasTeacherName := false
	for _, f := range fields {
		if f.Name == "teacher.id" {
			hasTeacherID = true
		}
		if f.Name == "teacher.name" {
			hasTeacherName = true
		}
	}
	if !hasTeacherID || !hasTeacherName {
		t.Error("CreateExamRequest 应展开 teacher 为 teacher.id 和 teacher.name")
	}
}

func TestFlattenSchema_AllOf(t *testing.T) {
	// 测试 allOf 合并
	schema := map[string]interface{}{
		"allOf": []interface{}{
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{"type": "string"},
				},
			},
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"age": map[string]interface{}{"type": "integer"},
				},
			},
		},
	}

	doc := &OpenAPIDocument{
		Version: "3.0",
		Raw: map[string]interface{}{
			"components": map[string]interface{}{
				"schemas": map[string]interface{}{},
			},
		},
	}
	s := NewSummarizer(doc)
	fields := s.flattenSchema(schema, nil)

	if len(fields) != 2 {
		t.Fatalf("allOf 合并后应有 2 个字段，实际 %d", len(fields))
	}
}

// ===== 工具函数测试 =====

func TestExtractSchemaName(t *testing.T) {
	cases := []struct {
		ref      string
		expected string
	}{
		{"#/components/schemas/AbnormalDetail", "AbnormalDetail"},
		{"#/components/schemas/ExamPageResult", "ExamPageResult"},
		{"#/definitions/User", "User"},
		{"", ""},
		{"no-slash", "no-slash"},
	}
	for _, c := range cases {
		result := extractSchemaName(c.ref)
		if result != c.expected {
			t.Errorf("extractSchemaName(%q) = %q, 期望 %q", c.ref, result, c.expected)
		}
	}
}

func TestEscapeMarkdown(t *testing.T) {
	if escapeMarkdown("a|b") != "a\\|b" {
		t.Error("管道符应被转义")
	}
	if escapeMarkdown("hello\nworld") != "hello world" {
		t.Error("换行应被替换为空格")
	}
}

func TestContainsString(t *testing.T) {
	if !containsString([]string{"a", "b", "c"}, "b") {
		t.Error("应找到 b")
	}
	if containsString([]string{"a", "b", "c"}, "z") {
		t.Error("不应找到 z")
	}
	if containsString(nil, "a") {
		t.Error("nil 切片不应找到任何值")
	}
}
