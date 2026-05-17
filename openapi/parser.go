package openapi

import (
	"encoding/json"
	"errors"
)

func ParseDocument(data []byte) (map[string]any, error) {
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, errors.New("invalid JSON document")
	}

	if _, ok := doc["paths"]; !ok {
		return nil, errors.New("invalid OpenAPI document: missing paths")
	}

	if DetectVersion(doc) == "unknown" {
		return nil, errors.New("unsupported specification version: expected OpenAPI 3.x or Swagger 2.0")
	}

	return doc, nil
}

func DetectVersion(doc map[string]any) string {
	if openapiVersion, ok := doc["openapi"].(string); ok {
		if len(openapiVersion) >= 1 {
			return "3.0"
		}
	}

	if swaggerVersion, ok := doc["swagger"].(string); ok && swaggerVersion == "2.0" {
		return "2.0"
	}

	return "unknown"
}

func ExtractTags(doc map[string]any) []string {
	tagSet := map[string]struct{}{}

	if topTags, ok := doc["tags"].([]any); ok {
		for _, tag := range topTags {
			tagMap, ok := tag.(map[string]any)
			if !ok {
				continue
			}
			name, ok := tagMap["name"].(string)
			if !ok || name == "" {
				continue
			}
			tagSet[name] = struct{}{}
		}
	}

	paths, ok := doc["paths"].(map[string]any)
	if !ok {
		return sortedKeys(tagSet)
	}

	for _, pathItem := range paths {
		pathMap, ok := pathItem.(map[string]any)
		if !ok {
			continue
		}
		for _, method := range httpMethods {
			operation, ok := pathMap[method].(map[string]any)
			if !ok {
				continue
			}
			collectOperationTags(tagSet, operation)
		}
	}

	return sortedKeys(tagSet)
}
