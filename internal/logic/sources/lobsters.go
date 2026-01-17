package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"github.com/Numpkens/grip/internal/logic"
)

type Lobsters struct {
	Client  *http.Client
	BaseURL string 
}

func (l *Lobsters) Search(ctx context.Context, query string) ([]logic.Post, error) {
   
	apiURL := l.BaseURL
	if apiURL == "" {
		apiURL = "https://lobste.rs"
	}
    
	url := fmt.Sprintf("%s/t/%s.json", apiURL, query)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "GripAggregator/1.0 (+https://github.com/Numpkens/grip; numpkins1222@gmail.com)")

	resp, err := l.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

   
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
    }

	var apiResults []struct {
		Title       string `json:"title"`
		URL         string `json:"url"`
		PublishedAt string `json:"created_at"`
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
			Source:      "Lobsters",
			PublishedAt: parsedDate,
		})
	}
	return posts, nil
}