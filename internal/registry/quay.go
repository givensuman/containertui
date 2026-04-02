package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const quayAPIBase = "https://quay.io/api/v1"

// QuayClient is a minimal client for public Quay repository search.
type QuayClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewQuayClient creates a new Quay API client.
func NewQuayClient() *QuayClient {
	return &QuayClient{
		httpClient: &http.Client{Timeout: DefaultTimeout},
		baseURL:    quayAPIBase,
	}
}

type quaySearchResponse struct {
	Repositories []quayRepository `json:"repositories"`
}

type quayRepository struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
	StarCount   int    `json:"star_count"`

	Popularity float64 `json:"popularity"`
}

// Search queries Quay for repositories matching the query.
func (c *QuayClient) Search(ctx context.Context, query string, pageSize int) (SearchResponse, error) {
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}

	params := url.Values{}
	params.Set("query", query)
	params.Set("limit", fmt.Sprintf("%d", pageSize))

	endpoint := fmt.Sprintf("%s/find/repositories?%s", c.baseURL, params.Encode())

	var raw quaySearchResponse
	if err := c.doRequest(ctx, endpoint, &raw); err != nil {
		return SearchResponse{}, fmt.Errorf("failed to search quay repositories: %w", err)
	}

	results := make([]RegistryImage, 0, len(raw.Repositories))
	for _, repo := range raw.Repositories {
		fullName := fmt.Sprintf("quay.io/%s/%s", repo.Namespace, repo.Name)
		results = append(results, RegistryImage{
			RepoName:         fullName,
			ShortDescription: repo.Description,
			StarCount:        repo.StarCount,
			PullCount:        int64(repo.Popularity),
			IsOfficial:       false,
			IsAutomated:      false,
			Registry:         "quay",
		})
	}

	return SearchResponse{Count: len(results), Results: results}, nil
}

func (c *QuayClient) doRequest(ctx context.Context, endpoint string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}
