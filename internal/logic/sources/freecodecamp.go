package sources

import (
	"bytes"
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
	
	url := "https://gql.hashnode.com"

	
	jsonData := map[string]interface{}{
		"query": `
			query {
				publication(host: "freecodecamp.org/news") {
					posts(first: 10, filter: { tagSlugs: ["` + query + `"] }) {
						edges {
							node {
								title
								url
								publishedAt
							}
						}
					}
				}
			}`,
	}

	body, _ := json.Marshal(jsonData)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Printf("[DEBUG] FCC status code: %d for query: %s\n", resp.StatusCode, query)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}


	var response struct {
		Data struct {
			Publication struct {
				Posts struct {
					Edges []struct {
						Node struct {
							Title       string `json:"title"`
							URL         string `json:"url"`
							PublishedAt string `json:"publishedAt"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"posts"`
			} `json:"publication"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	fmt.Printf("[DEBUG] FCC parsed %d posts from API\n", len(response.Data.Publication.Posts.Edges))

	var posts []logic.Post
	for _, edge := range response.Data.Publication.Posts.Edges {
		parsedDate, _ := time.Parse(time.RFC3339, edge.Node.PublishedAt)
		posts = append(posts, logic.Post{
			Title:       edge.Node.Title,
			URL:         edge.Node.URL,
			Source:      "FreeCodeCamp",
			PublishedAt: parsedDate,
		})
	}

	fmt.Printf("[DEBUG] FCC (Hashnode) found %d posts\n", len(posts))
	return posts, nil
}