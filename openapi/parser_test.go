package openapi

import "testing"

func TestDetectVersionSupportsOpenAPI3AndSwagger2(t *testing.T) {
	if got := DetectVersion(map[string]any{"openapi": "3.0.3"}); got != "3.0" {
		t.Fatalf("expected 3.0, got %s", got)
	}

	if got := DetectVersion(map[string]any{"swagger": "2.0"}); got != "2.0" {
		t.Fatalf("expected 2.0, got %s", got)
	}

	if got := DetectVersion(map[string]any{"openapi": "2.0.0"}); got != "unknown" {
		t.Fatalf("expected unknown for non-3.x openapi, got %s", got)
	}
}
