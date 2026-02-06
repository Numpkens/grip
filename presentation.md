  
#   
#          GRIP: Logic-First Aggregator  
#      ------------------------------------  
#         		David Gagnon

---

# The Core Architecture  
### "Brain vs. Body"

> "The engine doesn't know about HTTP. It only knows about Data."

```go  
// internal/logic/engine.go  
type Post struct {  
    Title       string    `json:"title"`  
    URL         string    `json:"url"`  
    Source      string    `json:"source"`  
    PublishedAt time.Time `json:"published_at"`  
}

type Source interface {  
    Search(ctx context.Context, query string) ([]Post, error)  
}
```
---

# **Concurrency: Fan-Out Pattern**

### **Parallelizing 6 Sources**

```Go  
// internal/logic/engine.go  
func (e *Engine) Collect(ctx context.Context, query string) []Post {  
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)  
    defer cancel()

    resultsChan := make(chan []Post, len(e.Sources))  
    var wg sync.WaitGroup

    for _, s := range e.Sources {  
        wg.Add(1)  
        go func(src Source) {  
            defer wg.Done()  
            posts, _ := src.Search(ctx, query)  
            resultsChan <- posts  
        }(s)  
    }  
}
```

---

# **Algorithmic Edge: Min-Heap**

### **Efficiency: O(N log K)**

```Go  
// Maintaining a constant Top 20 leaderboard  
for _, p := range posts {  
    if h.Len() < 20 {  
        heap.Push(h, p)  
    } else if p.PublishedAt.After((*h)[0].PublishedAt) {  
        // Root (*h)[0] is the 'Min' (Oldest)  
        heap.Pop(h)  
        heap.Push(h, p)  
    }  
}```

* **Flat Memory:** Footprint never grows beyond 20 items.  
* **Comparison:** New posts only enter if newer than the "oldest."

---

# **Cross-Language Reliability**

### **Consuming the Go API in Nim**

```Nim  
import std/[httpclient, json]

# Proving the Headless contract  
let client = newHttpClient()  
let response = client.getContent("http://localhost:8080/")  
let data = parseJson(response)

for post in data:  
    echo post["title"].getStr(), " | ", post["source"].getStr()
```
* **Proof:** Decoupled architecture allows any language to join.

---

# **Production Readiness**

### **Multi-Head Delivery**

* **Web UI:** `cmd/grip-web`  
* **JSON API:** `cmd/grip-api` (Port 8080)  
* **CLI:** `cmd/grip-cli`  
* **TUI:** `cmd/grip-tui` (Bubble Tea/Lip Gloss)

```Bash  
# Multi-stage Docker build  
docker build -t grip .  
docker run -p 8080:8080 grip
```
---

# **Final Summary**

* **Decoupled:** The engine is purely logic-driven.  
* **Concurrent:** Fan-out reduces latency by >50%.  
* **Resilient:** Enforced by 2s context timeouts.

### **Questions?**

---

