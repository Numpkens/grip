package logic

import (
	"testing"
	"time"
)

type TestSource struct {
	Posts []Post
}

func (ts *TestSource) Search(query string) ([]Post, error) {
	return ts.Posts, nil
}

func TestFetchAllSorting(t *testing.T) {
	now := time.Now()
	older := now.Add(-24 * time.Hour)

	s1 := &TestSource{Posts: []Post{{Title: "Old Post", PublishedAt: older}}}
	s2 := &TestSource{Posts: []Post{{Title: "New Post", PublishedAt: now}}}

	engine := &Engine{Sources: []Source{s1, s2}}
	results := engine.FetchAll("test")

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].Title != "New Post" {
		t.Errorf("Sorting failed: expected 'New Post' first, got '%s'", results[0].Title)
	}
}
