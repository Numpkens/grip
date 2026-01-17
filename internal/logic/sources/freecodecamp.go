package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"github.com/Numpkens/grip/internal/logic"
)

type FreeCodeCamp struct {
	Client  *http.Client
	BaseURL string 
}

func (f *FreeCodeCamp) Search(ctx context.Context, query string) ([]logic.Post, error) {
	const apiKey = "2524043232675924738541916"
	apiURL := f.BaseURL
	if apiURL == "" {
		apiURL = "https://www.freecodecamp.org"
	}
    
	finalURL := fmt.Sprintf("%s/news/ghost/api/v3/content/posts/?key=%s&filter=tag:%s", apiURL, apiKey, query)

	req, err := http.NewRequestWithContext(ctx, "GET", finalURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "GripAggregator/1.0 (+https://github.com/Numpkens/grip; numpkins1222@gmail.com)")

	fmt.Printf("[DEBUG] FCC Fetching: %s\n", finalURL)
	resp, err := f.Client.Do(req)
	if err != nil {
		fmt.Printf("[DEBUG] FCC Request Error: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Printf("[DEBUG] FCC Status Code: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var apiResults struct {
		Posts []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			PublishedAt string `json:"published_at"`
		} `json:"posts"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResults); err != nil {
		return nil, err 
	}

	var posts []logic.Post
	for _, r := range apiResults.Posts {
		parsedDate, _ := time.Parse(time.RFC3339, r.PublishedAt)
		posts = append(posts, logic.Post{
			Title:       r.Title,
			URL:         r.URL,
			Source:      "FreeCodeCamp",
			PublishedAt: parsedDate,
		})
	}
	return posts, nil
}