# GRIP: Go Reader Interface Processor

GRIP is a headless, concurrent search engine built to aggregate developer blogs into a single stream. I built this with a "Logic-First" approach, meaning the engine’s brain is entirely separate from the UI. It doesn't care if it’s feeding a browser, a terminal, or another application.

## Why I Built This
Most aggregators are slow or locked into one specific front-end. I wanted something lightweight that followed proper Go patterns  using a nested structure to keep routing and data logic separate.

## Core Architecture

### 1. The Headless Engine (internal/logic)
The engine is the central brain. It’s source-agnostic, meaning it doesn't know about HTTP or HTML. It just takes a search string, manages the goroutines, and hands back a clean slice of results.

### 2. Strategy Pattern & Interfaces
I use a Source interface so the project can scale without a total rewrite. Whether a provider uses JSON, GraphQL, XML, or an RSS feed, I can just plug it in. I started with Dev.to, but once the interface logic was solid, adding sources like HackerNews and Lobste.rs became a simple two-line addition to the main engine.

### 3. Concurrency: Fan-Out / Fan-In
Processing searches sequentially was too slow (~1000ms). I moved to a Fan-Out pattern where every source gets its own goroutine managed by a sync.WaitGroup. This brought response times down from ~1000ms to sub-500ms (currently averaging 471ms), even with six active sources.

## Engineering Decisions

### Smart Sorting with Min-Heaps
To keep memory usage constant, I didn't just sort a massive slice. I used a Min-Heap for a "Top 20" leaderboard. We maintain exactly 20 items; as new posts come in, we only keep them if they are newer than the oldest item on the heap. This gives us O(N log K) efficiency.

### "Good Citizen" Networking & Ethics
I didn't want to build a "blind" crawler. 
* **Respecting Robots.txt:** Before adding sources like Boot.dev, I checked their robots.txt to ensure I wasn't violating any rules.
* **Identification:** I identify my crawler in the headers by sending my GitHub repo URL and email so admins know who is hitting their server.
* **Resilience:** I use context.WithTimeout to enforce a strict 2-second limit. This prevents one hanging API from stalling the whole app.

## Headless Proof: Multiple Entry Points
The decoupling is proven by the fact that I have three different "heads" using the exact same logic:
1. **Web (cmd/grip):** A card-view **UI** built with html/template and a Swagger-documented **API**.
2. **CLI (cmd/cli):** A terminal tool for searching directly from the command line.

## Documentation
Technical documentation for the internal logic and API is available through:
* **Internal Logic:** Comprehensive documentation of exported types and concurrency patterns is maintained via [pkgdocs](https://pkg.go.dev/github.com/Numpkens/grip/internal/logic).
* **API Reference:** When the web server is running, the Swagger UI is available at `/swagger/index.html` to test endpoints and view schemas.
* **Architecture:** For a deep dive into the concurrency model and the Min-Heap sorting logic, see ARCHITECTURE.md in the root directory.

## Lessons Learned
The project had some growing pains. I originally struggled with the nested structure vs a flat one, but moving to internal/logic was the right call for scalability. I also learned the hard way that missing a pointer in a function can result in changing a copy of a slice in memory rather than the actual results, which taught me to be a lot more careful with how I log empty returns.