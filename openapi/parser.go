package openapi

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"
)

// OpenAPIDocument represents both OpenAPI 3.0 and Swagger 2.0 documents
type OpenAPIDocument struct {
	Version string // "3.0" or "2.0"
	Raw     map[string]interface{}
}

// PathItem represents an API endpoint
type PathItem struct {
	Path   string
	Tags   []string
	Methods map[string]interface{}
}

// ParseOpenAPI fetches and parses OpenAPI document from URL
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

// detectOpenAPIVersion detects if document is OpenAPI 3.0 or Swagger 2.0
func detectOpenAPIVersion(doc map[string]interface{}) string {
	// Check for OpenAPI 3.0
	if openapi, ok := doc["openapi"].(string); ok {
		if len(openapi) > 0 && openapi[0:1] == "3" {
			return "3.0"
		}
	}
	// Default to 2.0 (Swagger)
	return "2.0"
}

// GetAllTags extracts all unique tags from the document
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

			// Check all HTTP methods
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

			// Check all HTTP methods
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

	// Convert to sorted slice
	tags := make([]string, 0, len(tagsMap))
	for tag := range tagsMap {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

// GetPathsByTags returns all paths that belong to selected tags
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

func tagsFromMap(m map[string]bool) []string {
	tags := make([]string, 0, len(m))
	for tag := range m {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}
