package registry

import (
	"encoding/json"
	"testing"
)

func TestMapQuaySearchResultsParsesRepositoryEntries(t *testing.T) {
	rawJSON := []byte(`{
		"results": [
			{
				"kind": "repository",
				"name": "nginx-unprivileged",
				"description": "Nginx image",
				"namespace": {"name": "nginx"}
			},
			{
				"kind": "organization",
				"name": "nginx"
			}
		]
	}`)

	var payload quaySearchResponse
	if err := json.Unmarshal(rawJSON, &payload); err != nil {
		t.Fatalf("failed to unmarshal fixture: %v", err)
	}

	results := mapQuaySearchResults(payload)
	if len(results) != 1 {
		t.Fatalf("expected 1 repository result, got %d", len(results))
	}

	if results[0].RepoName != "quay.io/nginx/nginx-unprivileged" {
		t.Fatalf("unexpected repo name: %s", results[0].RepoName)
	}

	if results[0].Registry != "quay" {
		t.Fatalf("expected registry 'quay', got %q", results[0].Registry)
	}
}
