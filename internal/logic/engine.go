// Package logic provides the core "brain" of the GRIP aggregator.
// It implements a headless engine that uses a Fan-Out pattern to query multiple 
// developer blog sources concurrently. Results are aggregated and sorted 
// using a Min-Heap to ensure a constant memory footprint while returning 
// only the 20 most recent posts.
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
	URL         string    `json:"url" example:"https://dev.to/user/post"`
	Source      string    `json:"source" example:"dev.to"`
	PublishedAt time.Time `json:"published_at" example:"2026-01-21T10:00:00Z"`
}
// Source defines the contract for adding new source providers.
type Source interface {
	
	Search(ctx context.Context, query string) ([]Post, error)
}

//resultsHeap implements heap.Interface to maintain a Top 20 list by date.
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
// Collect triggers a concurrent fan-out at all sources and enforces a 2 second timeout rule and returns the 20 most recent posts.
func (e *Engine) Collect(ctx context.Context, query string) []Post {
	// Set a hard deadline for the entire collection process
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	h := &resultsHeap{}
	heap.Init(h)
	
	// Buffered channel prevents worker goroutines from hanging if we exit early
	resultsChan := make(chan []Post, len(e.Sources))
	var wg sync.WaitGroup

	for _, s := range e.Sources {
		wg.Add(1)
		go func(src Source) {
			defer wg.Done()
			// The context is passed to the search to cancel network calls if timeout hits
			posts, err := src.Search(ctx, query)
			if err != nil {
				return
			}
			resultsChan <- posts
		}(s)
	}

	// This goroutine ensures the channel is closed so the loop can finish if all sources report in
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	finished := 0
Loop:
	for finished < len(e.Sources) {
		select {
		case posts, ok := <-resultsChan:
			if !ok {
				break Loop
			}
			finished++
			
			for _, p := range posts {
				if h.Len() < 20 {
					heap.Push(h, p)
				} else if p.PublishedAt.After((*h)[0].PublishedAt) {
					heap.Pop(h)
					heap.Push(h, p)
				}
			}

		case <-ctx.Done():
			// 2-second timeout hit! Break and return what we have so far
			break Loop
		}
	}

	// Drain heap into a sorted "newest first" slice
	final := make([]Post, h.Len())
	for i := h.Len() - 1; i >= 0; i-- {
		final[i] = heap.Pop(h).(Post)
	}
	return final
}

func NewEngine(source []Source) *Engine {
	return &Engine{
		Sources: source,
	}
}