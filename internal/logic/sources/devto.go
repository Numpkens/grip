package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"github.com/Numpkens/grip/internal/logic"
)

type DevTo struct {
	Client  *http.Client
	BaseURL string 
}

func (d *DevTo) Search(ctx context.Context, query string) ([]logic.Post, error) {
   
	apiURL := d.BaseURL
	if apiURL == "" {
		apiURL = "https://dev.to/api"
	}
    
	url := fmt.Sprintf("%s/articles?tag=%s", apiURL, query)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := d.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

   
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("devto api error: status %d", resp.StatusCode)
    }

	var apiResults []struct {
		Title       string `json:"title"`
		URL         string `json:"url"`
		PublishedAt string `json:"published_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResults); err != nil {
		return nil, err 
	}

	var posts []logic.Post
	for _, r := range apiResults {
		parsedDate, _ := time.Parse(time.RFC3339, r.PublishedAt)

		posts = append(posts, logic.Post{
			Title:       r.Title,
			URL:         r.URL,
			Source:      "Dev.to",
			PublishedAt: parsedDate,
		})
	}
	return posts, nil
}