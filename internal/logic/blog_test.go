package logic

import (
	"testing"
	"time"
	"fmt"
	"sort"
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

func TestBootDevDateParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
		shouldFail bool
	}{
		{
			name:  "Valid RFC1123Z",
			input: "Fri, 16 Jan 2026 13:29:15 +0000",
			expected: time.Date(2026, 1, 16, 13, 29, 15, 0, time.UTC),
			shouldFail: false,
		},
		{
			name:  "Invalid format - should fallback",
			input: "2026-01-16",
			expected: time.Unix(0, 0),
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedDate, err := time.Parse(time.RFC1123Z, tt.input)
			if err != nil && !tt.shouldFail {
				t.Errorf("Expected no error for %s, got %v", tt.input, err)
			}
			
			if err != nil {
				parsedDate = time.Unix(0, 0)
			}

			if !parsedDate.Equal(tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, parsedDate)
			}
		})
	}
}

func TestEngineSorting(t *testing.T) {
	now := time.Now()
	posts := []Post{
		{Title: "Old Post", PublishedAt: now.Add(-24 * time.Hour)},
		{Title: "New Post", PublishedAt: now},
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].PublishedAt.After(posts[j].PublishedAt)
	})

	if posts[0].Title != "New Post" {
		t.Errorf("Sorting failed: Expected 'New Post' first, got %s", posts[0].Title)
	}
}

func TestFetchAllTruncation(t *testing.T) {
	var manyPosts []Post
	for i := 0; i < 25; i++ {
		manyPosts = append(manyPosts, Post{
			Title:       fmt.Sprintf("Post %d", i),
			PublishedAt: time.Now().Add(time.Duration(i) * time.Minute),
		})
	}

	s1 := &TestSource{Posts: manyPosts}
	engine := &Engine{Sources: []Source{s1}}
	results := engine.FetchAll("test")

	if len(results) != 20 {
		t.Errorf("Truncation failed: expected 20 results, got %d", len(results))
	}
}

func BenchmarkFetchAll(b *testing.B) {
	s1 := &TestSource{Posts: []Post{{Title: "A", PublishedAt: time.Now()}}}
	engine := &Engine{Sources: []Source{s1, s1, s1, s1}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.FetchAll("test")
	}
}
