package sources

import (
	"context"
	"encoding/xml"
	"net/http"
	"strings"
	"time"
	"github.com/Numpkens/grip/internal/logic"
)

type BootDev struct {
	Client *http.Client
}

type bootRSS struct {
	Channel struct {
		Items []struct {
			Title   string `xml:"title"`
			Link    string `xml:"link"`
			PubDate string `xml:"pubDate"`
		} `xml:"item"`
	} `xml:"channel"`
}

func (b *BootDev) Search(ctx context.Context, query string) ([]logic.Post, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://blog.boot.dev/index.xml", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "GripAggregator/1.0")

	resp, err := b.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rss bootRSS
	if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
		return nil, err
	}

	var posts []logic.Post
	for _, item := range rss.Channel.Items {
		if strings.Contains(strings.ToLower(item.Title), strings.ToLower(query)) {
			
			parsedDate, _ := time.Parse(time.RFC1123, item.PubDate)
			posts = append(posts, logic.Post{
				Title:       item.Title,
				URL:         item.Link,
				Source:      "Boot.dev",
				PublishedAt: parsedDate,
			})
		}
	}
	return posts, nil
}