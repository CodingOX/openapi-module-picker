// Package openapi 提供 OpenAPI/Swagger 文档的摘要和结构化查询功能。
// 将原始 OpenAPI JSON 转换为 LLM 友好的结构化 Markdown 输出。
package openapi

import (
	"fmt"
	"sort"
	"strings"
)

// ---- 输出类型 ----

// ResolvedField 表示展开了 $ref 之后的扁平字段信息。
type ResolvedField struct {
	Name        string // 字段名，嵌套字段用点号分隔如 "teacher.name"
	Type        string // 类型描述，如 "string"、"integer(int64)"、"AbnormalDetail[]"
	Required    bool   // 是否必填
	Title       string // 中文标签（取自 schema 的 title 字段）
	Description string // 字段说明
}

// TagSummary 表示一个业务模块的概要信息。
type TagSummary struct {
	Name        string // 标签名
	Description string // 标签描述
	PathCount   int    // 该标签下的接口数
}

// SchemaCount 表示一个数据模型的引用统计。
type SchemaCount struct {
	Name        string // 模型名
	FieldCount  int    // 字段数
	RefCount    int    // 被接口引用的次数
	Description string // 模型描述（取自 schema 的 description 字段）
}

// ---- Summarizer ----

// Summarizer 提供 OpenAPI 文档的文本摘要功能。
// 将原始的 OpenAPI JSON 转换为 LLM 友好的结构化 Markdown 输出。
type Summarizer struct {
	doc         *OpenAPIDocument
	schemaCache map[string]map[string]interface{} // $ref → resolved schema 缓存
}

// NewSummarizer 基于已解析的 OpenAPI 文档创建摘要器。
func NewSummarizer(doc *OpenAPIDocument) *Summarizer {
	return &Summarizer{
		doc:         doc,
		schemaCache: make(map[string]map[string]interface{}),
	}
}

// httpMethods 按常规顺序排列的 HTTP 方法列表。
var httpMethods = []string{"get", "post", "put", "delete", "patch", "options", "head", "trace"}

// ---- 公共方法 ----

// GenerateSummary 生成 API 全局概览 Markdown。
// 包含：基本信息、业务模块（标签）、Top 数据模型。
func (s *Summarizer) GenerateSummary() string {
	var b strings.Builder

	title := s.getDocInfo("title")
	version := s.getDocInfo("version")
	b.WriteString(fmt.Sprintf("# API 概览: %s v%s\n\n", title, version))

	// 统计总接口数
	tagSummary := s.buildTagSummary()
	totalPaths := 0
	for _, ts := range tagSummary {
		totalPaths += ts.PathCount
	}

	b.WriteString(fmt.Sprintf("**接口总数：** %d\n\n", totalPaths))

	// 业务模块表格
	b.WriteString("## 业务模块\n\n")
	b.WriteString("| 标签 | 接口数 | 描述 |\n")
	b.WriteString("|------|--------|------|\n")
	for _, ts := range tagSummary {
		desc := ts.Description
		if desc == "" {
			desc = "-"
		}
		b.WriteString(fmt.Sprintf("| %s | %d | %s |\n", escapeMarkdown(ts.Name), ts.PathCount, escapeMarkdown(desc)))
	}

	// Top 数据模型
	schemas := s.getTopSchemas(10)
	if len(schemas) > 0 {
		b.WriteString("\n## 数据模型（Top 10）\n\n")
		b.WriteString("| 模型 | 字段数 | 被引用次数 |\n")
		b.WriteString("|------|--------|-----------|\n")
		for _, sc := range schemas {
			b.WriteString(fmt.Sprintf("| %s | %d | %d |\n", escapeMarkdown(sc.Name), sc.FieldCount, sc.RefCount))
		}
	}

	return b.String()
}

// ListByTags 按标签列出 API 接口清单，输出 Markdown。
// tags 为空时按所有标签分组列出全部接口。
func (s *Summarizer) ListByTags(tags []string) string {
	var b strings.Builder
	b.WriteString("# API 接口清单\n\n")

	// 确定要展示的标签列表
	var activeTags []string
	if len(tags) == 0 {
		// 收集所有标签
		tagSet := make(map[string]bool)
		paths, ok := s.doc.Raw["paths"].(map[string]interface{})
		if ok {
			for _, pathItem := range paths {
				item, _ := pathItem.(map[string]interface{})
				for _, method := range httpMethods {
					op, _ := item[method].(map[string]interface{})
					if op == nil {
						continue
					}
					s.collectTags(op, tagSet)
				}
			}
		}
		for t := range tagSet {
			activeTags = append(activeTags, t)
		}
		sort.Strings(activeTags)
	} else {
		activeTags = tags
	}

	if len(activeTags) == 0 {
		b.WriteString("文档中未找到任何标签。\n")
		return b.String()
	}

	// 按标签分组输出接口
	for _, tag := range activeTags {
		tagEndpoints := s.getEndpointsByTag(tag)
		if len(tagEndpoints) == 0 {
			continue
		}

		desc := s.tagDescription(tag)
		if desc != "" {
			b.WriteString(fmt.Sprintf("## %s — %s (%d 个接口)\n\n", escapeMarkdown(tag), escapeMarkdown(desc), len(tagEndpoints)))
		} else {
			b.WriteString(fmt.Sprintf("## %s (%d 个接口)\n\n", escapeMarkdown(tag), len(tagEndpoints)))
		}

		b.WriteString("| 方法 | 路径 | 摘要 |\n")
		b.WriteString("|------|------|------|\n")
		for _, ep := range tagEndpoints {
			b.WriteString(fmt.Sprintf("| %s | `%s` | %s |\n",
				strings.ToUpper(ep.Method),
				escapeMarkdown(ep.Path),
				escapeMarkdown(ep.Summary)))
		}

		// 收集涉及的数据模型
		models := s.collectEndpointModels(tagEndpoints)
		if len(models) > 0 {
			modelLinks := make([]string, len(models))
			for i, m := range models {
				modelLinks[i] = fmt.Sprintf("`%s`", m)
			}
			b.WriteString(fmt.Sprintf("\n**涉及数据模型：** %s\n\n", strings.Join(modelLinks, "、")))
		}
	}

	return b.String()
}

// DescribeEndpoint 输出单个接口的完整契约 Markdown。
// 包含：请求参数（path/query/header/body）、响应结构（按状态码）。
func (s *Summarizer) DescribeEndpoint(path, method string) string {
	operation, err := s.findOperation(path, method)
	if err != nil {
		return fmt.Sprintf("**错误：** %v\n", err)
	}

	var b strings.Builder
	summary := getString(operation, "summary")
	desc := getString(operation, "description")

	b.WriteString(fmt.Sprintf("# %s %s\n\n", strings.ToUpper(method), path))
	if summary != "" {
		b.WriteString(fmt.Sprintf("**摘要：** %s\n\n", escapeMarkdown(summary)))
	}
	if desc != "" {
		b.WriteString(fmt.Sprintf("**说明：** %s\n\n", escapeMarkdown(desc)))
	}

	// 请求参数（path / query / header）
	s.writeParameters(&b, operation)

	// 请求体
	s.writeRequestBody(&b, operation)

	// 响应
	s.writeResponses(&b, operation)

	return b.String()
}

// GetAllSchemas 返回文档中所有数据模型的统计信息。
// 包含 components/schemas（OpenAPI 3.x）或 definitions（Swagger 2.0）中的全部模型，
// 按被接口引用次数降序排列，未引用的模型 RefCount 为 0。
func (s *Summarizer) GetAllSchemas() []SchemaCount {
	// 获取所有 schema 名称
	schemaNames := s.getAllSchemaNames()
	if len(schemaNames) == 0 {
		return nil
	}

	// 引用计数
	refCounts := s.countSchemaRefs()

	result := make([]SchemaCount, 0, len(schemaNames))
	for _, name := range schemaNames {
		sc := SchemaCount{
			Name:       name,
			FieldCount: s.getFieldCount(name),
			RefCount:   refCounts[name],
		}

		// 从 schema 自身提取 description
		if schema, err := s.resolveRef(s.schemaRef(name)); err == nil {
			if desc, ok := schema["description"].(string); ok {
				sc.Description = desc
			}
		}
		result = append(result, sc)
	}

	// 按 RefCount 降序排列
	sort.Slice(result, func(i, j int) bool {
		if result[i].RefCount != result[j].RefCount {
			return result[i].RefCount > result[j].RefCount
		}
		return result[i].Name < result[j].Name
	})

	return result
}

// getAllSchemaNames 返回文档中所有 schema 名称（已排序）。
func (s *Summarizer) getAllSchemaNames() []string {
	var schemaMap map[string]interface{}

	if s.doc.Version == "2.0" {
		if defs, ok := s.doc.Raw["definitions"].(map[string]interface{}); ok {
			schemaMap = defs
		}
	} else {
		if components, ok := s.doc.Raw["components"].(map[string]interface{}); ok {
			if schemas, ok := components["schemas"].(map[string]interface{}); ok {
				schemaMap = schemas
			}
		}
	}

	if schemaMap == nil {
		return nil
	}

	names := make([]string, 0, len(schemaMap))
	for name := range schemaMap {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// DescribeSchema 返回指定数据模型的展开字段列表。
// name 为 schema 名称（如 "AbnormalDetail"），自动根据文档版本
// 处理 OpenAPI 3.x 的 #/components/schemas/ 和 Swagger 2.0 的 #/definitions/ 路径。
// 返回的字段列表已递归展开 $ref 嵌套引用。
func (s *Summarizer) DescribeSchema(name string) ([]ResolvedField, error) {
	ref := s.schemaRef(name)
	schema, err := s.resolveRef(ref)
	if err != nil {
		return nil, fmt.Errorf("数据模型不存在: %s", name)
	}
	return s.flattenSchema(schema, nil), nil
}

// ---- $ref 解析 ----

// schemaRef 根据文档版本返回正确的 $ref 前缀路径。
// OpenAPI 3.x 使用 "#/components/schemas/"，Swagger 2.0 使用 "#/definitions/"。
func (s *Summarizer) schemaRef(name string) string {
	if s.doc.Version == "2.0" {
		return "#/definitions/" + name
	}
	return "#/components/schemas/" + name
}

// resolveRef 解析 $ref 字符串，返回引用的 schema 对象。
// 支持 "#/components/schemas/XXX" 等 JSON 指针格式。
// 内部维护缓存，重复解析直接返回缓存结果。
func (s *Summarizer) resolveRef(ref string) (map[string]interface{}, error) {
	if cached, ok := s.schemaCache[ref]; ok {
		return cached, nil
	}

	if !strings.HasPrefix(ref, "#/") {
		return nil, fmt.Errorf("不支持的外部 $ref: %s", ref)
	}

	// 去掉 "#/" 前缀，按 "/" 分割路径
	parts := strings.Split(strings.TrimPrefix(ref, "#/"), "/")
	current := s.doc.Raw
	for _, part := range parts {
		// JSON Pointer 转义还原
		key := strings.ReplaceAll(part, "~1", "/")
		key = strings.ReplaceAll(key, "~0", "~")
		if m, ok := current[key].(map[string]interface{}); ok {
			current = m
		} else {
			return nil, fmt.Errorf("无法解析 $ref: %s（在 %s 处断路）", ref, key)
		}
	}

	s.schemaCache[ref] = current
	return current, nil
}

// ---- Schema 扁平化 ----

// flattenSchema 将 schema 展开为扁平字段列表，递归展开 $ref。
// visited 用于防止循环引用。首次调用传 nil 即可。
func (s *Summarizer) flattenSchema(schema map[string]interface{}, visited map[string]bool) []ResolvedField {
	if visited == nil {
		visited = make(map[string]bool)
	}

	// 处理 $ref
	if ref, ok := schema["$ref"].(string); ok {
		if visited[ref] {
			return nil // 循环引用，跳过
		}
		visited[ref] = true
		defer delete(visited, ref)

		resolved, err := s.resolveRef(ref)
		if err != nil {
			return nil
		}
		return s.flattenSchema(resolved, visited)
	}

	// 处理 allOf（合并多个 schema）
	if allOf, ok := schema["allOf"].([]interface{}); ok {
		var fields []ResolvedField
		for _, item := range allOf {
			if itemMap, ok := item.(map[string]interface{}); ok {
				fields = append(fields, s.flattenSchema(itemMap, visited)...)
			}
		}
		return fields
	}

	// 处理 properties
	var fields []ResolvedField
	if props, ok := schema["properties"].(map[string]interface{}); ok {
		requiredList := s.requiredList(schema)

		// 字段名排序，保证输出稳定
		fieldNames := make([]string, 0, len(props))
		for name := range props {
			fieldNames = append(fieldNames, name)
		}
		sort.Strings(fieldNames)

		for _, name := range fieldNames {
			prop, ok := props[name].(map[string]interface{})
			if !ok {
				continue
			}

			field := ResolvedField{
				Name:        name,
				Required:    containsString(requiredList, name),
				Title:       getString(prop, "title"),
				Description: getString(prop, "description"),
			}

			// 属性为 $ref → 展开一级
			if ref, ok := prop["$ref"].(string); ok {
				if !visited[ref] {
					visited[ref] = true
					resolved, err := s.resolveRef(ref)
					if err == nil {
						subFields := s.flattenSchema(resolved, visited)
						if len(subFields) > 0 {
							for _, sf := range subFields {
								sf.Name = name + "." + sf.Name
								sf.Required = field.Required || sf.Required
								fields = append(fields, sf)
							}
						} else {
							// 展开了但没有字段（如循环引用或空对象），回退
							field.Type = extractSchemaName(ref)
							fields = append(fields, field)
						}
					} else {
						field.Type = extractSchemaName(ref)
						fields = append(fields, field)
					}
					delete(visited, ref)
				}
				continue
			}

			// 数组类型
			if getString(prop, "type") == "array" {
				field.Type = s.formatArrayType(prop)
			} else {
				field.Type = s.formatType(prop)
			}

			fields = append(fields, field)
		}
	}

	return fields
}

// getEndpointsByTag 返回属于指定标签的所有接口。
func (s *Summarizer) getEndpointsByTag(tag string) []EndpointInfo {
	var result []EndpointInfo
	paths, ok := s.doc.Raw["paths"].(map[string]interface{})
	if !ok {
		return result
	}

	// 路径排序
	pathNames := make([]string, 0, len(paths))
	for p := range paths {
		pathNames = append(pathNames, p)
	}
	sort.Strings(pathNames)

	for _, pathName := range pathNames {
		pathItem := paths[pathName].(map[string]interface{})
		for _, method := range httpMethods {
			op, ok := pathItem[method].(map[string]interface{})
			if !ok {
				continue
			}
			tags, _ := op["tags"].([]interface{})
			for _, t := range tags {
				if tStr, ok := t.(string); ok && tStr == tag {
					summary, _ := op["summary"].(string)
					result = append(result, EndpointInfo{
						Method:  method,
						Path:    pathName,
						Summary: summary,
					})
					break
				}
			}
		}
	}
	return result
}

// collectTags 从 operation 中收集所有 tag 到 set 中。
func (s *Summarizer) collectTags(op map[string]interface{}, set map[string]bool) {
	tags, _ := op["tags"].([]interface{})
	for _, t := range tags {
		if tStr, ok := t.(string); ok {
			set[tStr] = true
		}
	}
}

// collectEndpointModels 收集一组 endpoint 涉及的数据模型名称。
func (s *Summarizer) collectEndpointModels(endpoints []EndpointInfo) []string {
	modelSet := make(map[string]bool)
	for _, ep := range endpoints {
		paths, ok := s.doc.Raw["paths"].(map[string]interface{})
		if !ok {
			continue
		}
		pathItem, ok := paths[ep.Path].(map[string]interface{})
		if !ok {
			continue
		}
		op, ok := pathItem[ep.Method].(map[string]interface{})
		if !ok {
			continue
		}
		s.collectSchemaRefs(op, modelSet)
	}

	models := make([]string, 0, len(modelSet))
	for m := range modelSet {
		models = append(models, m)
	}
	sort.Strings(models)
	return models
}

// collectSchemaRefs 递归遍历 JSON 树，收集所有 $ref 中的 schema 名称。
func (s *Summarizer) collectSchemaRefs(node interface{}, set map[string]bool) {
	switch v := node.(type) {
	case map[string]interface{}:
		if ref, ok := v["$ref"].(string); ok {
			if name := extractSchemaName(ref); name != "" {
				set[name] = true
			}
		}
		for _, val := range v {
			s.collectSchemaRefs(val, set)
		}
	case []interface{}:
		for _, item := range v {
			s.collectSchemaRefs(item, set)
		}
	}
}

// buildTagSummary 收集每个标签的接口数和描述。
func (s *Summarizer) buildTagSummary() []TagSummary {
	tagMap := make(map[string]int)

	// 从 operations 中统计
	paths, ok := s.doc.Raw["paths"].(map[string]interface{})
	if !ok {
		return nil
	}
	for _, pathItem := range paths {
		item, _ := pathItem.(map[string]interface{})
		for _, method := range httpMethods {
			op, _ := item[method].(map[string]interface{})
			if op == nil {
				continue
			}
			tags, _ := op["tags"].([]interface{})
			for _, t := range tags {
				if tStr, ok := t.(string); ok {
					tagMap[tStr]++
				}
			}
		}
	}

	result := make([]TagSummary, 0, len(tagMap))
	for tag, count := range tagMap {
		result = append(result, TagSummary{
			Name:      tag,
			PathCount: count,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	// 从顶层 tags 数组中补充描述
	for i := range result {
		result[i].Description = s.tagDescription(result[i].Name)
	}

	return result
}

// tagDescription 从文档的顶层 tags 数组中查找标签描述。
func (s *Summarizer) tagDescription(tagName string) string {
	if rawTags, ok := s.doc.Raw["tags"].([]interface{}); ok {
		for _, raw := range rawTags {
			if tagObj, ok := raw.(map[string]interface{}); ok {
				if name, ok := tagObj["name"].(string); ok && name == tagName {
					if desc, ok := tagObj["description"].(string); ok {
						return desc
					}
				}
			}
		}
	}
	return ""
}

// getTopSchemas 返回按引用次数降序的 Top N 数据模型。
func (s *Summarizer) getTopSchemas(n int) []SchemaCount {
	refCounts := s.countSchemaRefs()
	if len(refCounts) == 0 {
		return nil
	}

	// 转换为切片并排序
	schemas := make([]SchemaCount, 0, len(refCounts))
	for name, count := range refCounts {
		schemas = append(schemas, SchemaCount{
			Name:       name,
			FieldCount: s.getFieldCount(name),
			RefCount:   count,
		})
	}
	sort.Slice(schemas, func(i, j int) bool {
		if schemas[i].RefCount != schemas[j].RefCount {
			return schemas[i].RefCount > schemas[j].RefCount
		}
		return schemas[i].Name < schemas[j].Name
	})

	if n > 0 && n < len(schemas) {
		schemas = schemas[:n]
	}
	return schemas
}

// countSchemaRefs 遍历文档中所有 operations 的 $ref，统计每个 schema 被引用次数。
func (s *Summarizer) countSchemaRefs() map[string]int {
	counts := make(map[string]int)
	paths, ok := s.doc.Raw["paths"].(map[string]interface{})
	if !ok {
		return counts
	}

	for _, pathItem := range paths {
		item, _ := pathItem.(map[string]interface{})
		for _, method := range httpMethods {
			op, _ := item[method].(map[string]interface{})
			if op == nil {
				continue
			}
			// 收集 schemas
			refSet := make(map[string]bool)
			s.collectSchemaRefs(op, refSet)
			for name := range refSet {
				counts[name]++
			}
		}
	}
	return counts
}

// getFieldCount 返回指定 schema 的字段数（顶级 properties 数量）。
func (s *Summarizer) getFieldCount(schemaName string) int {
	schema, err := s.resolveRef(s.schemaRef(schemaName))
	if err != nil {
		return 0
	}
	props, _ := schema["properties"].(map[string]interface{})
	return len(props)
}

// ---- 参数输出 ----

// writeParameters 输出 path/query/header 参数表格。
func (s *Summarizer) writeParameters(b *strings.Builder, operation map[string]interface{}) {
	rawParams, _ := operation["parameters"].([]interface{})
	if len(rawParams) == 0 {
		return
	}

	// 按位置分组
	groups := map[string][]map[string]interface{}{
		"path":   {},
		"query":  {},
		"header": {},
	}
	order := []string{"path", "query", "header"}
	groupLabels := map[string]string{
		"path":   "Path 参数",
		"query":  "Query 参数",
		"header": "Header 参数",
	}

	for _, raw := range rawParams {
		p, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		in, _ := p["in"].(string)
		if _, exists := groups[in]; exists {
			groups[in] = append(groups[in], p)
		}
	}

	hasAny := false
	for _, loc := range order {
		params := groups[loc]
		if len(params) == 0 {
			continue
		}
		hasAny = true
		b.WriteString(fmt.Sprintf("### %s\n\n", groupLabels[loc]))
		b.WriteString("| 参数名 | 类型 | 必填 | 说明 |\n")
		b.WriteString("|--------|------|------|------|\n")

		for _, p := range params {
			name, _ := p["name"].(string)
			required, _ := p["required"].(bool)
			desc, _ := p["description"].(string)
			typeStr := s.paramType(p)
			reqStr := "否"
			if required {
				reqStr = "是"
			}
			if desc == "" {
				desc = "-"
			}
			b.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s |\n",
				escapeMarkdown(name), typeStr, reqStr, escapeMarkdown(desc)))
		}
		b.WriteString("\n")
	}

	if !hasAny {
		// 参数都在不支持的 "in" 位置（如 cookie）
	}
}

// paramType 提取参数的 schema 类型。
func (s *Summarizer) paramType(param map[string]interface{}) string {
	if schema, ok := param["schema"].(map[string]interface{}); ok {
		return s.formatType(schema)
	}
	return "string"
}

// writeRequestBody 输出请求体 Markdown 表格。
func (s *Summarizer) writeRequestBody(b *strings.Builder, operation map[string]interface{}) {
	rb, ok := operation["requestBody"].(map[string]interface{})
	if !ok {
		return
	}

	required, _ := rb["required"].(bool)
	reqStr := "可选"
	if required {
		reqStr = "必填"
	}

	content, _ := rb["content"].(map[string]interface{})
	if content == nil || len(content) == 0 {
		return
	}

	// 取第一个 content type（通常是 application/json）
	var contentType string
	var schema map[string]interface{}
	for ct, c := range content {
		contentType = ct
		if cObj, ok := c.(map[string]interface{}); ok {
			if s, ok := cObj["schema"].(map[string]interface{}); ok {
				schema = s
			}
		}
		break
	}

	bodyLabel := "Body"
	if contentType != "" {
		bodyLabel = fmt.Sprintf("Body（%s）", contentType)
	}

	if schema == nil {
		b.WriteString(fmt.Sprintf("### %s\n\n请求体已定义但无 schema 信息。\n\n", bodyLabel))
		return
	}

	fields := s.flattenSchema(schema, nil)
	if len(fields) == 0 {
		b.WriteString(fmt.Sprintf("### %s（%s）\n\n请求体无展开字段。\n\n", bodyLabel, reqStr))
		return
	}

	b.WriteString(fmt.Sprintf("### %s（%s）\n\n", bodyLabel, reqStr))
	b.WriteString("| 字段 | 类型 | 必填 | 说明 |\n")
	b.WriteString("|------|------|------|------|\n")
	for _, f := range fields {
		req := "否"
		if f.Required {
			req = "是"
		}
		desc := f.Title
		if f.Description != "" {
			desc = f.Description
		}
		if desc == "" {
			desc = "-"
		}
		b.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s |\n",
			escapeMarkdown(f.Name), f.Type, req, escapeMarkdown(desc)))
	}
	b.WriteString("\n")
}

// writeResponses 输出响应 Markdown。
func (s *Summarizer) writeResponses(b *strings.Builder, operation map[string]interface{}) {
	responses, ok := operation["responses"].(map[string]interface{})
	if !ok || len(responses) == 0 {
		return
	}

	b.WriteString("## 响应\n\n")

	// 状态码排序
	statusCodes := make([]string, 0, len(responses))
	for code := range responses {
		statusCodes = append(statusCodes, code)
	}
	sort.Strings(statusCodes)

	for _, code := range statusCodes {
		resp, _ := responses[code].(map[string]interface{})
		if resp == nil {
			continue
		}

		respDesc, _ := resp["description"].(string)
		if code == "default" {
			b.WriteString(fmt.Sprintf("### 默认响应 — %s\n\n", escapeMarkdown(respDesc)))
		} else {
			b.WriteString(fmt.Sprintf("### %s — %s\n\n", code, escapeMarkdown(respDesc)))
		}

		content, _ := resp["content"].(map[string]interface{})
		if content == nil || len(content) == 0 {
			b.WriteString("无响应体。\n\n")
			continue
		}

		// 取第一个 content type
		var schema map[string]interface{}
		for _, c := range content {
			if cObj, ok := c.(map[string]interface{}); ok {
				if s, ok := cObj["schema"].(map[string]interface{}); ok {
					schema = s
				}
			}
			break
		}

		if schema == nil {
			continue
		}

		fields := s.flattenSchema(schema, nil)
		if len(fields) == 0 {
			continue
		}

		b.WriteString("| 字段 | 类型 | 说明 |\n")
		b.WriteString("|------|------|------|\n")
		for _, f := range fields {
			desc := f.Title
			if f.Description != "" {
				desc = f.Description
			}
			if desc == "" {
				desc = "-"
			}
			b.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n",
				escapeMarkdown(f.Name), f.Type, escapeMarkdown(desc)))
		}
		b.WriteString("\n")
	}
}

// ---- 辅助方法 ----

// getDocInfo 从 info 对象中获取指定字段的值。
func (s *Summarizer) getDocInfo(field string) string {
	info, ok := s.doc.Raw["info"].(map[string]interface{})
	if !ok {
		return ""
	}
	val, _ := info[field].(string)
	return val
}

// findOperation 根据路径和方法名定位 operation。
func (s *Summarizer) findOperation(path, method string) (map[string]interface{}, error) {
	paths, ok := s.doc.Raw["paths"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("文档中无 paths 字段")
	}
	pathItem, ok := paths[path].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("路径不存在: %s", path)
	}
	operation, ok := pathItem[method].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("路径 %s 下无 %s 方法", path, strings.ToUpper(method))
	}
	return operation, nil
}

// formatType 格式化类型描述，如 "integer(int64)"、"string"。
func (s *Summarizer) formatType(schema map[string]interface{}) string {
	t := getString(schema, "type")
	format := getString(schema, "format")
	if t == "" {
		// 没有 type 但有 enum 的，归类为 string
		if _, ok := schema["enum"]; ok {
			return "string(enum)"
		}
		// 可能是一个引用类型（在属性中已经处理了）
		return "any"
	}
	if format != "" {
		return fmt.Sprintf("%s(%s)", t, format)
	}
	return t
}

// formatArrayType 格式化数组类型，如 "string[]"、"AbnormalDetail[]"。
func (s *Summarizer) formatArrayType(schema map[string]interface{}) string {
	items, ok := schema["items"].(map[string]interface{})
	if !ok {
		return "array"
	}
	// items 为 $ref
	if ref, ok := items["$ref"].(string); ok {
		return extractSchemaName(ref) + "[]"
	}
	// items 为内联类型
	t := s.formatType(items)
	return t + "[]"
}

// requiredList 从 schema 中提取 required 数组。
func (s *Summarizer) requiredList(schema map[string]interface{}) []string {
	raw, _ := schema["required"].([]interface{})
	list := make([]string, 0, len(raw))
	for _, r := range raw {
		if s, ok := r.(string); ok {
			list = append(list, s)
		}
	}
	return list
}

// ---- 工具函数 ----

// getString 安全地从 map 中提取 string 值。
func getString(m map[string]interface{}, key string) string {
	v, _ := m[key].(string)
	return v
}

// containsString 检查字符串切片中是否包含指定值。
func containsString(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

// extractSchemaName 从 $ref 字符串中提取 schema 名称。
// "#/components/schemas/AbnormalDetail" → "AbnormalDetail"
func extractSchemaName(ref string) string {
	parts := strings.Split(ref, "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

// escapeMarkdown 对 Markdown 表格中的特殊字符进行转义。
func escapeMarkdown(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

// EscapeMarkdown 是 escapeMarkdown 的公开包装，供 cmd 层使用。
func EscapeMarkdown(s string) string {
	return escapeMarkdown(s)
}
