package logic

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type TestSource struct {
	Posts []Post
}

func (s *TestSource) Search(ctx context.Context, query string) ([]Post, error) {
	return s.Posts, nil
}

func TestEngine_Collect_HeapLogic(t *testing.T) {
	now := time.Now()
	var manyPosts []Post
	for i := 0; i < 25; i++ {
		manyPosts = append(manyPosts, Post{
			Title:       fmt.Sprintf("Post %d", i),
			PublishedAt: now.Add(time.Duration(i) * time.Hour),
		})
	}

	s1 := &TestSource{Posts: manyPosts}
	engine := &Engine{Sources: []Source{s1}}

	results := engine.Collect(context.Background(), "test")

	if len(results) != 20 {
		t.Fatalf("expected 20 results, got %d", len(results) )
	}

	for _, p := range results {
		if p.Title == "Post 0" || p.Title == "Post 1" {
			t.Errorf("heap failed to evict oldest posts: found %s", p.Title)
		}
	}
}

func TestEngine_Collect_Concurrency(t *testing.T) {
	now := time.Now()
	s1 := &TestSource{Posts: []Post{{Title: "S1", PublishedAt: now}}}
	s2 := &TestSource{Posts: []Post{{Title: "S2", PublishedAt: now.Add(time.Hour)}}}

	engine := &Engine{Sources: []Source{s1, s2}}
	results := engine.Collect(context.Background(), "test")

	if len(results) != 2 {
		t.Errorf("expected 2 results from concurrent sources, got %d", len(results))
	}
}

func TestBootDevDateParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		{
			input:    "Fri, 16 Jan 2026 13:29:15 +0000",
			expected: time.Date(2026, 1, 16, 13, 29, 15, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parsedDate, err := time.Parse(time.RFC1123Z, tt.input)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}
			if !parsedDate.Equal(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, parsedDate)
			}
		})
	}
}