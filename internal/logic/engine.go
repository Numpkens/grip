// Package logic provides the core "brain" of the GRIP aggregator.
//
// It implements a headless engine that uses a Fan-Out pattern to query multiple 
// developer blog sources concurrently. Results are aggregated and sorted 
// using a Min-Heap to ensure only the top 20 newest posts are returned.
//
// See ARCHITECTURE.md in the project root for a deep dive into the 
// concurrency model and source-agnostic design.
package logic

import (
	"container/heap"
	"context"
	"sync"
	"time"
)
// Post represents a standardized blog post from any external source.
type Post struct {
	Title       string    `json:"title" example:"golang"`
	URL         string    `json:"url" example: "https://dev.to/user/post"`
	Source      string    `json:"source" example: "dev.to"`
	PublishedAt time.Time `json:"published_at" example: "2026-01-21T10:00:00Z"`
}
// Source defines the contract for adding new source providers.
// Any struct implementing Search can be added to the engine.
type Source interface {
	
	Search(ctx context.Context, query string) ([]Post, error)
}


type resultsHeap []Post

func (h resultsHeap) Len() int           { return len(h) }
func (h resultsHeap) Less(i, j int) bool { return h[i].PublishedAt.Before(h[j].PublishedAt) }
func (h resultsHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *resultsHeap) Push(x interface{}) { *h = append(*h, x.(Post)) }
func (h *resultsHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type Engine struct {
	Sources []Source
}
// Collect triggers a concurrent fan-out at all sources and enforces a 2 second timeout rule.
// Returns the 20 newest posts
func (e *Engine) Collect (ctx context.Context, query string) []Post {
	
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	h := &resultsHeap{}
	heap.Init(h)
	
	resultsChan := make(chan []Post, len(e.Sources))
	var wg sync.WaitGroup

	for _, s := range e.Sources {
		wg.Add(1)
		go func(src Source) {
			defer wg.Done()
	
			posts, err := src.Search(ctx, query)
			if err != nil {
				return
			}
			resultsChan <- posts
		}(s)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for posts := range resultsChan {
		for _, p := range posts {
			if h.Len() < 20 {
				heap.Push(h, p)
			} else if p.PublishedAt.After((*h)[0].PublishedAt) {
				heap.Pop(h)
				heap.Push(h, p)
			}
		}
	}

	result := make([]Post, h.Len())
	for i := h.Len() - 1; i >= 0; i-- {
		result[i] = heap.Pop(h).(Post)
	}
	return result
}

func NewEngine(source []Source) *Engine {
	return &Engine{
		Sources: source,
	}
}