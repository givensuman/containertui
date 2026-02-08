package registry

import (
	"context"
	"testing"
	"time"
)

func TestDockerHubClient(t *testing.T) {
	client := NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("GetPopularImages", func(t *testing.T) {
		images, err := client.GetPopularImages(ctx, 5)
		if err != nil {
			t.Fatalf("GetPopularImages failed: %v", err)
		}

		if len(images) == 0 {
			t.Error("Expected at least one popular image")
		}

		// Check first image has required fields
		if len(images) > 0 {
			img := images[0]
			if img.RepoName == "" {
				t.Error("Expected RepoName to be non-empty")
			}
			if img.PullCount == 0 {
				t.Error("Expected PullCount to be non-zero for popular image")
			}
		}
	})

	t.Run("Search", func(t *testing.T) {
		response, err := client.Search(ctx, "nginx", 5)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if len(response.Results) == 0 {
			t.Error("Expected at least one search result for 'nginx'")
		}

		// Verify nginx is in results
		found := false
		for _, img := range response.Results {
			if img.RepoName == "nginx" {
				found = true
				if !img.IsOfficial {
					t.Error("Expected official nginx image to have IsOfficial=true")
				}
				break
			}
		}
		if !found {
			t.Error("Expected to find 'nginx' in search results")
		}
	})

	t.Run("GetRepository", func(t *testing.T) {
		detail, err := client.GetRepository(ctx, "library", "nginx")
		if err != nil {
			t.Fatalf("GetRepository failed: %v", err)
		}

		if detail.Name != "nginx" {
			t.Errorf("Expected name 'nginx', got '%s'", detail.Name)
		}

		if detail.Namespace != "library" {
			t.Errorf("Expected namespace 'library', got '%s'", detail.Namespace)
		}

		if detail.FullDescription == "" {
			t.Error("Expected FullDescription to be non-empty")
		}
	})
}
