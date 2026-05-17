package openapi

import "testing"

func TestExtractTagsIncludesTopLevelAndOperations(t *testing.T) {
	doc := map[string]any{
		"openapi": "3.0.0",
		"tags": []any{
			map[string]any{"name": "users"},
			map[string]any{"name": "reports"},
		},
		"paths": map[string]any{
			"/users": map[string]any{
				"get": map[string]any{"tags": []any{"users"}},
			},
			"/orders": map[string]any{
				"post": map[string]any{"tags": []any{"orders"}},
			},
		},
	}

	tags := ExtractTags(doc)
	if len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %d (%v)", len(tags), tags)
	}
}

func TestFilterByTagsFiltersOperationsAndPreservesOriginal(t *testing.T) {
	original := map[string]any{
		"swagger": "2.0",
		"paths": map[string]any{
			"/users": map[string]any{
				"get":  map[string]any{"tags": []any{"users"}, "summary": "list users"},
				"post": map[string]any{"tags": []any{"admin"}, "summary": "create user"},
			},
			"/orders": map[string]any{
				"get": map[string]any{"tags": []any{"orders"}, "summary": "list orders"},
			},
		},
		"tags": []any{
			map[string]any{"name": "users"},
			map[string]any{"name": "orders"},
			map[string]any{"name": "admin"},
		},
	}

	filtered, err := FilterByTags(original, []string{"users"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	paths := filtered["paths"].(map[string]any)
	if len(paths) != 1 {
		t.Fatalf("expected 1 path after filtering, got %d", len(paths))
	}

	usersPath := paths["/users"].(map[string]any)
	if _, ok := usersPath["get"]; !ok {
		t.Fatal("expected users GET to remain")
	}
	if _, ok := usersPath["post"]; ok {
		t.Fatal("expected users POST to be removed")
	}

	originalPaths := original["paths"].(map[string]any)
	originalUsersPath := originalPaths["/users"].(map[string]any)
	if _, ok := originalUsersPath["post"]; !ok {
		t.Fatal("expected original document to remain unchanged")
	}

	tags := filtered["tags"].([]any)
	if len(tags) != 1 {
		t.Fatalf("expected filtered top-level tags to include only selected tag, got %d", len(tags))
	}
}
