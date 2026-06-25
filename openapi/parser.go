// Package openapi 提供 OpenAPI/Swagger 文档的解析和过滤功能。
// 支持 OpenAPI 3.x 和 Swagger 2.0 规范，可从远程 URL 获取文档并按标签筛选 API 路径。
package openapi

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sort"
)

// OpenAPIDocument 表示 OpenAPI 3.0 或 Swagger 2.0 文档。
// 包含文档版本号和原始 JSON 数据，用于后续的标签提取和过滤操作。
type OpenAPIDocument struct {
	Version string                 // 文档版本："3.0" 或 "2.0"
	Raw     map[string]interface{} // 原始 JSON 数据
}

// PathItem 表示一个 API 端点，包含路径、关联的标签和 HTTP 方法。
type PathItem struct {
	Path    string                 // API 路径，如 "/users"
	Tags    []string               // 该路径关联的标签列表
	Methods map[string]interface{} // HTTP 方法到操作定义的映射
}

// EndpointInfo 表示一个 API 端点的基本信息，供 Summarizer 和 CLI 层使用。
type EndpointInfo struct {
	Method  string   // HTTP 方法，如 "get", "post"
	Path    string   // API 路径，如 "/users"
	Summary string   // 接口摘要
	Tags    []string // 关联标签列表
}

// ParseOpenAPI 从指定 URL 获取并解析 OpenAPI 文档。
// 自动检测文档版本（OpenAPI 3.x 或 Swagger 2.0），返回解析后的文档对象。
// 如果 URL 无效或文档格式错误，返回相应的错误。
func ParseOpenAPI(url string) (*OpenAPIDocument, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var doc map[string]interface{}
	err = json.Unmarshal(body, &doc)
	if err != nil {
		return nil, err
	}

	version := detectOpenAPIVersion(doc)
	return &OpenAPIDocument{
		Version: version,
		Raw:     doc,
	}, nil
}

// ParseOpenAPIFromFile 从本地文件解析 OpenAPI 文档。
// 自动检测文档版本（OpenAPI 3.x 或 Swagger 2.0）。
func ParseOpenAPIFromFile(path string) (*OpenAPIDocument, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc map[string]interface{}
	err = json.Unmarshal(data, &doc)
	if err != nil {
		return nil, err
	}
	version := detectOpenAPIVersion(doc)
	return &OpenAPIDocument{Version: version, Raw: doc}, nil
}

// detectOpenAPIVersion 检测文档是 OpenAPI 3.x 还是 Swagger 2.0。
// 通过检查 JSON 中的 "openapi" 字段判断版本：
//   - 如果 "openapi" 字段存在且以 "3" 开头，返回 "3.0"
//   - 否则默认返回 "2.0"（Swagger）
func detectOpenAPIVersion(doc map[string]interface{}) string {
	// 检查 OpenAPI 3.x 版本
	if openapi, ok := doc["openapi"].(string); ok {
		if len(openapi) > 0 && openapi[0:1] == "3" {
			return "3.0"
		}
	}
	// 默认为 Swagger 2.0
	return "2.0"
}

// GetAllTags 提取文档中所有唯一的标签。
// 遍历所有路径和 HTTP 方法，收集并去重标签列表，返回排序后的结果。
// 支持 OpenAPI 3.x 和 Swagger 2.0 格式。
func (doc *OpenAPIDocument) GetAllTags() []string {
	tagsMap := make(map[string]bool)

	if doc.Version == "3.0" {
		paths, ok := doc.Raw["paths"].(map[string]interface{})
		if !ok {
			return []string{}
		}

		for _, pathItem := range paths {
			item, ok := pathItem.(map[string]interface{})
			if !ok {
				continue
			}

			// 检查所有 HTTP 方法
			for _, method := range []string{"get", "post", "put", "delete", "patch", "options", "head", "trace"} {
				operation, ok := item[method].(map[string]interface{})
				if !ok {
					continue
				}

				if tagsList, ok := operation["tags"].([]interface{}); ok {
					for _, tag := range tagsList {
						if tagStr, ok := tag.(string); ok && tagStr != "" {
							tagsMap[tagStr] = true
						}
					}
				}
			}
		}
	} else { // Swagger 2.0
		paths, ok := doc.Raw["paths"].(map[string]interface{})
		if !ok {
			return []string{}
		}

		for _, pathItem := range paths {
			item, ok := pathItem.(map[string]interface{})
			if !ok {
				continue
			}

			// 检查所有 HTTP 方法
			for _, method := range []string{"get", "post", "put", "delete", "patch", "options", "head"} {
				operation, ok := item[method].(map[string]interface{})
				if !ok {
					continue
				}

				if tagsList, ok := operation["tags"].([]interface{}); ok {
					for _, tag := range tagsList {
						if tagStr, ok := tag.(string); ok && tagStr != "" {
							tagsMap[tagStr] = true
						}
					}
				}
			}
		}
	}

	// 转换为排序后的切片
	tags := make([]string, 0, len(tagsMap))
	for tag := range tagsMap {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

// GetPathsByTags 返回所有属于选定标签的 API 路径。
// 遍历文档中的所有路径，检查每个路径是否包含选定的标签。
// 如果路径的任意操作包含选定标签，则保留整个路径。
// 返回的路径列表按路径名排序。
func (doc *OpenAPIDocument) GetPathsByTags(selectedTags map[string]bool) []PathItem {
	var paths []PathItem

	if doc.Version == "3.0" {
		pathsMap, ok := doc.Raw["paths"].(map[string]interface{})
		if !ok {
			return paths
		}

		for pathName, pathItem := range pathsMap {
			item, ok := pathItem.(map[string]interface{})
			if !ok {
				continue
			}

			tagsSet := make(map[string]bool)
			methodsMap := make(map[string]interface{})

			for _, method := range []string{"get", "post", "put", "delete", "patch", "options", "head", "trace"} {
				operation, ok := item[method].(map[string]interface{})
				if !ok {
					continue
				}

				if tagsList, ok := operation["tags"].([]interface{}); ok {
					for _, tag := range tagsList {
						if tagStr, ok := tag.(string); ok {
							tagsSet[tagStr] = true
						}
					}
				}
				methodsMap[method] = operation
			}

			// Check if any tag in this path is selected
			for tag := range tagsSet {
				if selectedTags[tag] {
					paths = append(paths, PathItem{
						Path:    pathName,
						Tags:    tagsFromMap(tagsSet),
						Methods: methodsMap,
					})
					break
				}
			}
		}
	} else { // Swagger 2.0
		pathsMap, ok := doc.Raw["paths"].(map[string]interface{})
		if !ok {
			return paths
		}

		for pathName, pathItem := range pathsMap {
			item, ok := pathItem.(map[string]interface{})
			if !ok {
				continue
			}

			tagsSet := make(map[string]bool)
			methodsMap := make(map[string]interface{})

			for _, method := range []string{"get", "post", "put", "delete", "patch", "options", "head"} {
				operation, ok := item[method].(map[string]interface{})
				if !ok {
					continue
				}

				if tagsList, ok := operation["tags"].([]interface{}); ok {
					for _, tag := range tagsList {
						if tagStr, ok := tag.(string); ok {
							tagsSet[tagStr] = true
						}
					}
				}
				methodsMap[method] = operation
			}

			// Check if any tag in this path is selected
			for tag := range tagsSet {
				if selectedTags[tag] {
					paths = append(paths, PathItem{
						Path:    pathName,
						Tags:    tagsFromMap(tagsSet),
						Methods: methodsMap,
					})
					break
				}
			}
		}
	}

	return paths
}

// GetAllEndpoints 返回文档中所有 API 端点的扁平列表，不按标签分组。
// 按路径名和方法名排序，支持按 HTTP 方法过滤（通过 CLI 层实现）。
func (doc *OpenAPIDocument) GetAllEndpoints() []EndpointInfo {
	var result []EndpointInfo
	paths, ok := doc.Raw["paths"].(map[string]interface{})
	if !ok {
		return result
	}

	pathNames := make([]string, 0, len(paths))
	for p := range paths {
		pathNames = append(pathNames, p)
	}
	sort.Strings(pathNames)

	for _, pathName := range pathNames {
		pathItem, _ := paths[pathName].(map[string]interface{})
		for _, method := range []string{"get", "post", "put", "delete", "patch", "options", "head", "trace"} {
			op, ok := pathItem[method].(map[string]interface{})
			if !ok {
				continue
			}
			summary, _ := op["summary"].(string)
			var tags []string
			if tagsList, ok := op["tags"].([]interface{}); ok {
				for _, t := range tagsList {
					if tStr, ok := t.(string); ok {
						tags = append(tags, tStr)
					}
				}
			}
			result = append(result, EndpointInfo{
				Method:  method,
				Path:    pathName,
				Summary: summary,
				Tags:    tags,
			})
		}
	}
	return result
}

// tagsFromMap 将标签映射转换为排序后的字符串切片。
// 用于内部辅助，确保标签列表的一致性。
func tagsFromMap(m map[string]bool) []string {
	tags := make([]string, 0, len(m))
	for tag := range m {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}
