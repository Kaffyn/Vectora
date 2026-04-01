package tools

import (
	"context"
	"fmt"
	"net/http"
	"io"
)

// SearchResult representa um achado na web.
type SearchResult struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Snippet string `json:"snippet"`
}

// SearchService centralizes external searches and data scraping.
type SearchService struct {
	client *http.Client
}

func NewSearchService() *SearchService {
	return &SearchService{client: &http.Client{}}
}

// WebSearch executes a search (e.g., via Google/Serper/Bing).
func (s *SearchService) WebSearch(ctx context.Context, query string) ([]SearchResult, error) {
	return nil, fmt.Errorf("search_api_not_configured")
}

// FetchURL scrapes page text to feed the LLM.
func (s *SearchService) FetchURL(ctx context.Context, url string) (string, error) {
	resp, err := s.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
