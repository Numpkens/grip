# Architecture: GRIP Aggregator

## Overview
GRIP is a **headless, concurrent search engine** designed to aggregate developer blog content from multiple sources (Dev.to, Lobsters, etc.) into a single stream sorted by date published. The system is built with a "Logic-First" approach, ensuring the core engine is entirely decoupled from the delivery method whether that be web, external call or CLI.

## Core Components

### 1. The Headless Engine (`internal/logic`)
The Engine serves as the central brain. It is "source-agnostic," meaning it has no knowledge of HTTP, HTML, or JSON. It simply accepts a context.Context and a search string, and returns a slice.

### 2. Source Agnosticism (The Strategy Pattern)
The system defines a Source interface. Any new blog or API can be added to GRIP without modifying any of the engine code.
* **Implementation:** Individual sources (e.g., sources.DevTo, sources.Lobsters) are defined in internal/logic/sources.
* **Dependency Injection:** Sources are instantiated and injected into the engine at the application entry point (main.go).

### 3. Concurrency Model: Fan-Out / Fan-In
To ensure high performance and low latency, GRIP uses a **Fan-Out** pattern:
* The Engine spawns a separate **Goroutine** for every registered source.
* Each source performs its network I/O independently.
* A sync.WaitGroup ensures the Engine waits for all sources to report back (or timeout) before proceeding.

## Technical Design Decisions

### Min-Heap Sorting ($O(N \log K)$)
Instead of collecting thousands of results and performing a heavy sort on a massive slice, GRIP utilizes a **Min-Heap** (resultsHeap) to maintain a "Top 20" list in real-time.
* **Why:** This is to ensure the memory footprint remains constant regardless of how many results the external APIs return.
* **Efficiency:** We only keep the 20 most recent items, popping the oldest items off the heap as newer ones are found.

### Resilience: Context & Timeouts
Grip uses strict lifecycle management:
* The Engine enforces a **2-second timeout** using `context.WithTimeout`.
* Using .WithTimeout prevents a single slow or unresponsive API (e.g., a service outage at Dev.to) from hanging the entire application.
* **Request Chaining:** The web handler passes the request context (r.Context()) to the engine, ensuring that if a user cancels the request, all underlying API calls are immediately aborted.

## Decoupled Delivery (Headless Proof)
The project demonstrates true decoupling by providing two distinct entry points that share the same business logic:

1. **Web Head (cmd/grip):** Serves a UI using Go templates and provides a RESTful JSON API documented via Swagger.
2. **CLI Head (cmd/cli):** A lightweight terminal tool that executes searches and streams results directly to stdout.

## Data Flow
1. **Entry:** An entry point (cmd/grip or cmd/cli) initializes the http.Client and engine.
2. **Execution:** The Engine calls Collect(), fanning out goroutines to each registered source.
3. **Collection:** Results are sent back over a channel and pushed into the Min-Heap.
4. **Finalization:** The Engine returns the sorted slice. 
5. **Presentation:** * The **Web Handler** executes an HTML template or encodes JSON.
    * The **CLI** iterates and prints to the terminal.