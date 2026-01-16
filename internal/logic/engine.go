package logic

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

type Post struct {
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Source      string    `json:"source"`
	PublishedAt time.Time `json:"published_at"`
}

type Source interface {
	
	Search(ctx context.Context, query string) ([]Post, error)
}


type PostHeap []Post

func (h PostHeap) Len() int           { return len(h) }
func (h PostHeap) Less(i, j int) bool { return h[i].PublishedAt.Before(h[j].PublishedAt) }
func (h PostHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *PostHeap) Push(x interface{}) { *h = append(*h, x.(Post)) }
func (h *PostHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type Engine struct {
	Sources []Source
}

func (e *Engine) FetchAll(query string) []Post {
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	h := &PostHeap{}
	heap.Init(h)
	
	var mu sync.Mutex
	var wg sync.WaitGroup
	resultsChan := make(chan []Post, len(e.Sources))

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
		mu.Lock()
		for _, p := range posts {
			if h.Len() < 20 {
				heap.Push(h, p)
			} else if p.PublishedAt.After((*h)[0].PublishedAt) {
	
				heap.Pop(h)
				heap.Push(h, p)
			}
		}
		mu.Unlock()
	}

	
	result := make([]Post, h.Len())
	for i := h.Len() - 1; i >= 0; i-- {
		result[i] = heap.Pop(h).(Post)
	}

	return result
}