package openapi

import (
	"encoding/json"
)

// FilterByTags creates a new OpenAPI document containing only selected tags
// It preserves all content without modification, only removes unselected paths
func (doc *OpenAPIDocument) FilterByTags(selectedTags []string) ([]byte, error) {
	// Create a copy of the original document
	filtered := make(map[string]interface{})
	
	// Deep copy the document
	data, err := json.Marshal(doc.Raw)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &filtered)
	if err != nil {
		return nil, err
	}

	// Convert selected tags to map for quick lookup
	selectedTagsMap := make(map[string]bool)
	for _, tag := range selectedTags {
		selectedTagsMap[tag] = true
	}

	// Filter paths
	if paths, ok := filtered["paths"].(map[string]interface{}); ok {
		pathsToRemove := []string{}

		for pathName, pathItem := range paths {
			item, ok := pathItem.(map[string]interface{})
			if !ok {
				continue
			}

			// Check all HTTP methods to see if any has selected tags
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

			// Mark path for removal if it doesn't have selected tags
			if !hasSelectedTag {
				pathsToRemove = append(pathsToRemove, pathName)
			}
		}

		// Remove unmarked paths
		for _, pathName := range pathsToRemove {
			delete(paths, pathName)
		}
	}

	// Convert back to JSON
	result, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return nil, err
	}

	return result, nil
}
