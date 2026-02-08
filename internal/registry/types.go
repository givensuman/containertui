// Package registry provides Docker Hub API client functionality.
package registry

// RegistryImage represents a Docker Hub image in search results.
type RegistryImage struct {
	RepoName         string `json:"repo_name"`
	ShortDescription string `json:"short_description"`
	StarCount        int    `json:"star_count"`
	PullCount        int64  `json:"pull_count"`
	IsOfficial       bool   `json:"is_official"`
	IsAutomated      bool   `json:"is_automated"`
}

// RegistryImageDetail represents detailed information from Docker Hub.
type RegistryImageDetail struct {
	User            string `json:"user"`
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
	Description     string `json:"description"`
	FullDescription string `json:"full_description"` // README content
	StarCount       int    `json:"star_count"`
	PullCount       int64  `json:"pull_count"`
	LastUpdated     string `json:"last_updated"`
	DateRegistered  string `json:"date_registered"`
	IsPrivate       bool   `json:"is_private"`
	IsOfficial      bool   `json:"is_official"`
}

// SearchResponse represents the response from Docker Hub search API.
type SearchResponse struct {
	Count    int             `json:"count"`
	Next     string          `json:"next"`
	Previous string          `json:"previous"`
	Results  []RegistryImage `json:"results"`
}

// RepositoryListResponse represents the response from repository list API.
type RepositoryListResponse struct {
	Count    int                  `json:"count"`
	Next     string               `json:"next"`
	Previous string               `json:"previous"`
	Results  []RepositoryListItem `json:"results"`
}

// RepositoryListItem represents an item in the repository list response.
type RepositoryListItem struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
	StarCount   int    `json:"star_count"`
	PullCount   int64  `json:"pull_count"`
	LastUpdated string `json:"last_updated"`
}
