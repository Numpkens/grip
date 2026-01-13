package logic

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Source interface {
	Search(query string) ([]Post, error)
}

type DevTo struct{}

func (d *DevTo) Search(query string) ([]Post, error) {
	url := fmt.Sprintf("https://dev.to/api/articles?tag=%s", query)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("devto api call failed: %w", err)
	}

	defer resp.Body.Close()

	var apiResults []struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResults); err != nil {
		return nil, fmt.Errorf("failed to decode devto json: %w", err)
	}

	var posts []Post
	for _, r := range apiResults {
		posts = append(posts, Post{
			Title:  r.Title,
			URL:    r.URL,
			Source: "Dev.to",
		})
	}
	return posts, nil
}

type Post struct {
	Title  string
	URL    string
	Source string
}

type Engine struct {
	Sources []Source
}

func (e *Engine) FetchAll(query string) []Post {
	var allPosts []Post
	for _, s := range e.Sources {
		posts, err := s.Search(query)
		if err != nil {
			fmt.Printf("Error fetching from source: %v\n", err)
			continue
		}
		allPosts = append(allPosts, posts...)
	}
	return allPosts
}
