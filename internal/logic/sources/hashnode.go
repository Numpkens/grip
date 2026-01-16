package sources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
	"github.com/Numpkens/grip/internal/logic"
)

type Hashnode struct {
	Client *http.Client
}

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

func (h *Hashnode) Search(ctx context.Context, query string) ([]logic.Post, error) {
	tagSlug := slugify(query)

	queryStr := fmt.Sprintf(`{
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
    }`, tagSlug)

	jsonData := map[string]string{"query": queryStr}
	body, _ := json.Marshal(jsonData)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://gql.hashnode.com", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Tag struct {
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
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var posts []logic.Post
	for _, edge := range result.Data.Tag.Posts.Edges {
		parsedDate, _ := time.Parse(time.RFC3339, edge.Node.PublishedAt)
		posts = append(posts, logic.Post{
			Title:       edge.Node.Title,
			URL:         edge.Node.URL,
			Source:      "Hashnode",
			PublishedAt: parsedDate,
		})
	}
	return posts, nil
}