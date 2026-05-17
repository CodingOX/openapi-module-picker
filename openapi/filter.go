package openapi

import (
	"encoding/json"
)

var httpMethods = []string{"get", "post", "put", "delete", "patch", "head", "options", "trace"}

func FilterByTags(original map[string]any, selectedTags []string) (map[string]any, error) {
	selected := make(map[string]struct{}, len(selectedTags))
	for _, tag := range selectedTags {
		if tag == "" {
			continue
		}
		selected[tag] = struct{}{}
	}

	copied, err := deepCopy(original)
	if err != nil {
		return nil, err
	}

	paths, ok := copied["paths"].(map[string]any)
	if !ok {
		return copied, nil
	}

	for pathKey, pathItem := range paths {
		pathMap, ok := pathItem.(map[string]any)
		if !ok {
			delete(paths, pathKey)
			continue
		}

		keepPath := false
		for _, method := range httpMethods {
			op, exists := pathMap[method]
			if !exists {
				continue
			}
			operation, ok := op.(map[string]any)
			if !ok || !operationMatchesSelectedTags(operation, selected) {
				delete(pathMap, method)
				continue
			}
			keepPath = true
		}

		if !keepPath {
			delete(paths, pathKey)
		}
	}

	if tags, ok := copied["tags"].([]any); ok {
		filteredTags := make([]any, 0, len(tags))
		for _, tag := range tags {
			tagMap, ok := tag.(map[string]any)
			if !ok {
				continue
			}
			name, ok := tagMap["name"].(string)
			if !ok {
				continue
			}
			if _, exists := selected[name]; exists {
				filteredTags = append(filteredTags, tag)
			}
		}
		copied["tags"] = filteredTags
	}

	return copied, nil
}

func operationMatchesSelectedTags(operation map[string]any, selected map[string]struct{}) bool {
	tags, ok := operation["tags"].([]any)
	if !ok || len(tags) == 0 {
		return false
	}

	for _, rawTag := range tags {
		tagName, ok := rawTag.(string)
		if !ok {
			continue
		}
		if _, exists := selected[tagName]; exists {
			return true
		}
	}

	return false
}

func collectOperationTags(tagSet map[string]struct{}, operation map[string]any) {
	tags, ok := operation["tags"].([]any)
	if !ok {
		return
	}

	for _, rawTag := range tags {
		tagName, ok := rawTag.(string)
		if !ok || tagName == "" {
			continue
		}
		tagSet[tagName] = struct{}{}
	}
}

func deepCopy(source map[string]any) (map[string]any, error) {
	data, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}

	var target map[string]any
	if err := json.Unmarshal(data, &target); err != nil {
		return nil, err
	}

	return target, nil
}
