package openapi

import (
	"encoding/json"
	"strings"
)

// FilterByTags 创建一个仅包含选定标签的新 OpenAPI 文档。
// 通过深拷贝原始文档，移除不包含选定标签的路径，保留所有其他内容不变。
// 过滤粒度为路径级别：只要路径下任意操作的标签命中选定标签，即保留整个路径。
// 返回格式化的 JSON 字节数据，如果处理失败返回错误。
func (doc *OpenAPIDocument) FilterByTags(selectedTags []string) ([]byte, error) {
	// 创建原始文档的副本
	filtered := make(map[string]interface{})

	// 深拷贝文档，确保原始文档不被修改
	data, err := json.Marshal(doc.Raw)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &filtered)
	if err != nil {
		return nil, err
	}

	// 将选定标签转换为映射，便于快速查找
	selectedTagsMap := make(map[string]bool)
	for _, tag := range selectedTags {
		selectedTagsMap[tag] = true
	}

	// 过滤路径
	if paths, ok := filtered["paths"].(map[string]interface{}); ok {
		pathsToRemove := []string{}

		for pathName, pathItem := range paths {
			item, ok := pathItem.(map[string]interface{})
			if !ok {
				continue
			}

			// 检查所有 HTTP 方法，看是否有任何操作包含选定标签
			hasSelectedTag := false
			methodsToCheck := []string{"get", "post", "put", "delete", "patch", "options", "head", "trace"}

			for _, method := range methodsToCheck {
				operation, ok := item[method].(map[string]interface{})
				if !ok {
					continue
				}

				if tagsList, ok := operation["tags"].([]interface{}); ok {
					for _, tag := range tagsList {
						if tagStr, ok := tag.(string); ok && selectedTagsMap[tagStr] {
							hasSelectedTag = true
							break
						}
					}
					if hasSelectedTag {
						break
					}
				}
			}

			// 标记需要移除的路径（不包含选定标签）
			if !hasSelectedTag {
				pathsToRemove = append(pathsToRemove, pathName)
			}
		}

		// 移除未标记的路径
		for _, pathName := range pathsToRemove {
			delete(paths, pathName)
		}
	}

	// 同步过滤顶层 tags 数组，只保留选中标签的元数据条目
	if rawTags, ok := filtered["tags"].([]interface{}); ok {
		filteredTags := make([]interface{}, 0, len(rawTags))
		for _, tag := range rawTags {
			if tagObj, ok := tag.(map[string]interface{}); ok {
				if name, ok := tagObj["name"].(string); ok && selectedTagsMap[name] {
					filteredTags = append(filteredTags, tag)
				}
			} else if tagStr, ok := tag.(string); ok && selectedTagsMap[tagStr] {
				// 部分文档 tags 为纯字符串数组
				filteredTags = append(filteredTags, tag)
			}
		}
		filtered["tags"] = filteredTags
	}

	// 裁剪 components/schemas（OpenAPI 3）或 definitions（Swagger 2）中未被引用的 schema
	pruneUnreferencedSchemas(filtered, doc.Version)

	// 转换回 JSON 格式
	result, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ---- Schema 裁剪辅助函数 ----

// collectRefs 递归遍历任意 JSON 结构，收集所有 $ref 的值。
func collectRefs(v interface{}) []string {
	refs := make(map[string]bool)
	collectRefsRecursive(v, refs)
	result := make([]string, 0, len(refs))
	for ref := range refs {
		result = append(result, ref)
	}
	return result
}

// collectRefsRecursive 是 collectRefs 的递归实现。
func collectRefsRecursive(v interface{}, refs map[string]bool) {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, child := range val {
			if k == "$ref" {
				if refStr, ok := child.(string); ok {
					refs[refStr] = true
				}
			} else {
				collectRefsRecursive(child, refs)
			}
		}
	case []interface{}:
		for _, child := range val {
			collectRefsRecursive(child, refs)
		}
	}
}

// parseSchemaRef 从 $ref 字符串中提取 schema 名称。
// 支持格式：#/components/schemas/Foo 和 #/definitions/Foo，
// 包括嵌套指针如 #/components/schemas/Foo/properties/bar。
// 不匹配时返回空字符串。
func parseSchemaRef(ref string) string {
	prefixes := []string{"#/components/schemas/", "#/definitions/"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(ref, prefix) {
			name := strings.TrimPrefix(ref, prefix)
			if idx := strings.Index(name, "/"); idx != -1 {
				name = name[:idx]
			}
			return name
		}
	}
	return ""
}

// pruneUnreferencedSchemas 移除 components/schemas（OpenAPI 3）或
// definitions（Swagger 2）中未被 paths 引用的 schema。
// 计算传递闭包：若 schema A 引用 schema B，则 B 也应保留。
func pruneUnreferencedSchemas(filtered map[string]interface{}, version string) {
	var schemas map[string]interface{}
	var ok bool

	if version == "3.0" {
		components, has := filtered["components"].(map[string]interface{})
		if !has {
			return
		}
		schemas, ok = components["schemas"].(map[string]interface{})
		if !ok || len(schemas) == 0 {
			return
		}
	} else {
		schemas, ok = filtered["definitions"].(map[string]interface{})
		if !ok || len(schemas) == 0 {
			return
		}
	}

	// paths 为空或不存在时，移除所有 schema
	paths, hasPaths := filtered["paths"].(map[string]interface{})
	if !hasPaths || len(paths) == 0 {
		for name := range schemas {
			delete(schemas, name)
		}
		return
	}

	// 从 paths 中收集所有 $ref，计算传递闭包
	refs := collectRefs(paths)
	kept := transitiveSchemaClosure(refs, schemas)

	// 移除未引用的 schema
	for name := range schemas {
		if !kept[name] {
			delete(schemas, name)
		}
	}
}

// transitiveSchemaClosure BFS 计算 schema 引用的传递闭包。
// 从初始 $ref 集合出发，解析出 schema 名称，递归收集被引用的子 schema。
// 返回所有应保留的 schema 名称集合。
func transitiveSchemaClosure(initialRefs []string, schemas map[string]interface{}) map[string]bool {
	kept := make(map[string]bool)
	queue := make([]string, 0)

	// 第一轮：从初始 refs 中提取 schema 名
	for _, ref := range initialRefs {
		if name := parseSchemaRef(ref); name != "" {
			if !kept[name] {
				kept[name] = true
				queue = append(queue, name)
			}
		}
	}

	// BFS：递归收集被引用的子 schema
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]

		if schema, exists := schemas[name]; exists && schema != nil {
			refs := collectRefs(schema)
			for _, ref := range refs {
				if nestedName := parseSchemaRef(ref); nestedName != "" {
					if !kept[nestedName] {
						kept[nestedName] = true
						queue = append(queue, nestedName)
					}
				}
			}
		}
	}

	return kept
}
