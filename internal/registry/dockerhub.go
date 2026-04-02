package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	// DockerHubAPIBase is the base URL for Docker Hub v2 API.
	DockerHubAPIBase = "https://hub.docker.com/v2"

	// DefaultPageSize is the default number of results per page.
	DefaultPageSize = 25

	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 30 * time.Second
)

// Client is a Docker Hub API client.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new Docker Hub API client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		baseURL: DockerHubAPIBase,
	}
}

// Search queries Docker Hub for images matching the given query.
func (c *Client) Search(ctx context.Context, query string, pageSize int) (SearchResponse, error) {
	if pageSize == 0 {
		pageSize = DefaultPageSize
	}

	params := url.Values{}
	params.Set("query", query)
	params.Set("page_size", fmt.Sprintf("%d", pageSize))

	endpoint := fmt.Sprintf("%s/search/repositories/?%s", c.baseURL, params.Encode())

	var response SearchResponse
	if err := c.doRequest(ctx, endpoint, &response); err != nil {
		return SearchResponse{}, fmt.Errorf("failed to search repositories: %w", err)
	}

	for i := range response.Results {
		response.Results[i].Registry = "dockerhub"
	}

	return response, nil
}

// GetRepository fetches detailed information for a specific repository.
func (c *Client) GetRepository(ctx context.Context, namespace, name string) (RegistryImageDetail, error) {
	endpoint := fmt.Sprintf("%s/repositories/%s/%s/", c.baseURL, namespace, name)

	var detail RegistryImageDetail
	if err := c.doRequest(ctx, endpoint, &detail); err != nil {
		return RegistryImageDetail{}, fmt.Errorf("failed to get repository: %w", err)
	}

	return detail, nil
}

// GetPopularImages fetches popular/official images from Docker Hub.
// This returns official images from the "library" namespace sorted by pull count.
func (c *Client) GetPopularImages(ctx context.Context, pageSize int) ([]RegistryImage, error) {
	if pageSize == 0 {
		pageSize = 50
	}

	params := url.Values{}
	params.Set("page_size", fmt.Sprintf("%d", pageSize))
	params.Set("ordering", "-pull_count")

	endpoint := fmt.Sprintf("%s/repositories/library/?%s", c.baseURL, params.Encode())

	var response RepositoryListResponse
	if err := c.doRequest(ctx, endpoint, &response); err != nil {
		return nil, fmt.Errorf("failed to get popular images: %w", err)
	}

	// Convert RepositoryListItem to RegistryImage
	images := make([]RegistryImage, 0, len(response.Results))
	for _, item := range response.Results {
		images = append(images, RegistryImage{
			RepoName:         item.Name,
			ShortDescription: item.Description,
			StarCount:        item.StarCount,
			PullCount:        item.PullCount,
			IsOfficial:       item.Namespace == "library",
			IsAutomated:      false,
			Registry:         "dockerhub",
		})
	}

	return images, nil
}

// doRequest performs an HTTP GET request and unmarshals the JSON response.
func (c *Client) doRequest(ctx context.Context, endpoint string, result interface{}) error {
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
