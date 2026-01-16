package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Numpkens/grip/internal/logic"
)

type HackerNews struct {
	Client *http.Client
}

func (h *HackerNews) Search(ctx context.Context, query string) ([]logic.Post, error) {
	
	url := fmt.Sprintf("https://hn.algolia.com/api/v1/search?query=%s&tags=story", query)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := h.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Hits []struct {
			Title     string    `json:"title"`
			URL       string    `json:"url"`
			CreatedAt time.Time `json:"created_at"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var posts []logic.Post
	for _, hit := range result.Hits {
		posts = append(posts, logic.Post{
			Title:       hit.Title,
			URL:         hit.URL,
			Source:      "Hacker News",
			PublishedAt: hit.CreatedAt,
		})
	}
	return posts, nil
}