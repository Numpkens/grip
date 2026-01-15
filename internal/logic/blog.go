package logic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type Source interface {
	Search(query string) ([]Post, error)
}

type Hashnode struct{}

func slugify(query string) string {
	s := strings.ToLower(strings.TrimSpace(query))
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	reg := regexp.MustCompile(`[^a-z0-9-]+`)
	s = reg.ReplaceAllString(s, "")
	regMulti := regexp.MustCompile(`-+/`)
	s = regMulti.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func (h *Hashnode) Search(query string) ([]Post, error) {
	tagSlug := slugify(query)

	jsonData := map[string]string{
		"query": fmt.Sprintf(`{
            tag(slug: "%s") {
                posts(first: 10, filter: { sortBy: recent }) {
                    edges {
                        node {
                            title
                            url
                            publishedAt
                        }
                    }
                }
            }
        }`, tagSlug),
	}

	body, _ := json.Marshal(jsonData)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post("https://gql.hashnode.com", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResults struct {
		Data struct {
			Tag *struct {
				Posts struct {
					Edges []struct {
						Node struct {
							Title       string `json:"title"`
							URL         string `json:"url"`
							PublishedAt string `json:"publishedAt"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"posts"`
			} `json:"tag"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResults); err != nil {
		return nil, err
	}

	if len(apiResults.Errors) > 0 {
		return nil, fmt.Errorf("hashnode gql error: %s", apiResults.Errors[0].Message)
	}

	if apiResults.Data.Tag == nil {
		return []Post{}, nil
	}

	var posts []Post
	for _, edge := range apiResults.Data.Tag.Posts.Edges {
		n := edge.Node
		parsedDate, err := time.Parse(time.RFC3339, n.PublishedAt)
		if err != nil {
			parsedDate = time.Now()
		}

		posts = append(posts, Post{
			Title:       n.Title,
			URL:         n.URL,
			Source:      "Hashnode",
			PublishedAt: parsedDate,
		})
	}
	return posts, nil
}

type HackerNews struct{}

func (h *HackerNews) Search(query string) ([]Post, error) {
	url := fmt.Sprintf("https://hn.algolia.com/api/v1/search?query=%s&tags=story", query)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HackerNews API failed: %w", err)
	}
	defer resp.Body.Close()

	var apiResults struct {
		Hits []struct {
			Title     string `json:"title"`
			URL       string `json:"url"`
			CreatedAt string `json:"created_at"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResults); err != nil {
		return nil, fmt.Errorf("failed to decode HN json: %w", err)
	}

	var posts []Post
	for _, r := range apiResults.Hits {
		if r.URL == "" {
			continue
		}
		parsedDate, _ := time.Parse(time.RFC3339, r.CreatedAt)

		posts = append(posts, Post{
			Title:       r.Title,
			URL:         r.URL,
			Source:      "HackerNews",
			PublishedAt: parsedDate,
		})
	}
	return posts, nil
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
		Title       string `json:"title"`
		URL         string `json:"url"`
		PublishedAt string `json:"published_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResults); err != nil {
		return nil, fmt.Errorf("failed to decode devto json: %w", err)
	}

	var posts []Post
	for _, r := range apiResults {
		parsedDate, _ := time.Parse(time.RFC3339, r.PublishedAt)

		posts = append(posts, Post{
			Title:       r.Title,
			URL:         r.URL,
			Source:      "Dev.to",
			PublishedAt: parsedDate,
		})
	}
	return posts, nil
}

type Post struct {
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Source      string    `json:"source"`
	PublishedAt time.Time `json:"published_at"`
}

type Engine struct {
	Sources []Source
}

func (e *Engine) FetchAll(query string) []Post {
	var allPosts []Post
	var wg sync.WaitGroup
	resultsChan := make(chan []Post, len(e.Sources))

	for _, s := range e.Sources {
		wg.Add(1)
		go func(src Source) {
			defer wg.Done()
			posts, err := src.Search(query)
			if err != nil {
				fmt.Printf("Error fetching from source: %v\n", err)
				return
			}
			resultsChan <- posts
		}(s)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for posts := range resultsChan {
		allPosts = append(allPosts, posts...)
	}

	sort.Slice(allPosts, func(i, j int) bool {
		return allPosts[i].PublishedAt.After(allPosts[j].PublishedAt)
	})

	if len(allPosts) > 20 {
		allPosts = allPosts[:20]
	}
	return allPosts
}
