package openapi

import (
	"encoding/json"
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

	// 转换回 JSON 格式
	result, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return nil, err
	}

	return result, nil
}
