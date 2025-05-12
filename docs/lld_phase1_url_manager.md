# Low-Level Design (LLD): Phase 1 - URL Manager (Frontier)

## 1. Introduction

*   **Purpose:** This document details the Low-Level Design for the URL Manager (Frontier) component of the Phase 1 MVP web crawler.
*   **Scope:** Defines how URLs are queued, managed for duplicate prevention, and provided to Fetcher workers. For the MVP, this will be an in-memory implementation.
*   **Reference HLD:** `hld_phase1_mvp.md`

## 2. Responsibilities

*   Accept initial seed URLs.
*   Accept newly discovered URLs from Parser workers.
*   Maintain a queue of URLs pending to be crawled.
*   Provide the next available URL to a Fetcher worker when requested.
*   Prevent duplicate URLs from being added to the queue or processed multiple times within the same crawl session.
*   Track the number of URLs in the queue and potentially the number of URLs processed.
*   Signal when the frontier is empty and no more work is available (for graceful shutdown).

## 3. Interfaces

*   **Input Methods:**
    *   `AddSeeds(urls []string)`: Called once at startup to add initial URLs.
    *   `AddNewURLs(urls []string)`: Called by Parser workers to submit newly found links.
*   **Output Methods:**
    *   `GetNextURL() (url string, ok bool)`: Called by Fetcher workers. Returns the next URL to crawl and a boolean `ok` which is `false` if the queue is empty and no more URLs are expected (or a shutdown is signaled).
*   **Control/Status Methods (Optional for MVP, but good to consider):**
    *   `QueueSize() int`: Returns the current number of URLs in the queue.
    *   `IsDone() bool`: Indicates if crawling should stop (e.g., queue empty and workers idle).
*   **Internal Interaction:**
    *   Receives URLs from the main application (seeds) and Parser workers.
    *   Provides URLs to Fetcher workers.

## 4. Data Structures

```go
// Package: urlmanager
import (
    "sync"
)

type Manager struct {
    mu         sync.Mutex // Protects access to queue and visited
    queue      []string   // Simple slice acting as a FIFO queue for URLs
    visited    map[string]bool // Tracks URLs already added to queue or processed
    // For MVP, we might use a channel directly as the queue for simpler concurrency,
    // or a slice protected by a mutex as shown here.
    // Let's proceed with a channel-based approach for the queue for better concurrency handling.

    urlChan    chan string       // Buffered channel for URLs to be crawled
    // visited map still needed to avoid re-adding to urlChan

    // To manage shutdown gracefully
    activeWorkers sync.WaitGroup
    shutdownChan  chan struct{} // To signal workers to stop
    done          bool          // Flag to indicate all work is done
}
```

**Chosen Approach for MVP:**
*   `urlChan chan string`: A buffered channel will serve as the primary queue. Fetchers will read from this channel.
*   `visited map[string]bool`: A map protected by a mutex to track all URLs ever added to `urlChan` or processed, to prevent duplicates.
*   `mu sync.Mutex`: To protect the `visited` map and the `done` flag.
*   `shutdownChan chan struct{}`: To signal all operations to cease.
*   `activeTasks int`: (Protected by `mu`) A counter for active tasks (URLs sent to fetchers but not yet fully processed by parsers and storage) to help determine when the crawl is truly finished.

## 5. Core Logic / Algorithms

*   **Initialization (`NewManager` function):**
    1.  Initialize `urlChan` as a buffered channel (e.g., buffer size of `NumWorkers * 2`).
    2.  Initialize `visited` as an empty `map[string]bool`.
    3.  Initialize `shutdownChan`.
    4.  Return the new `Manager` instance.

*   **`AddURLs(urls []string, isSeed bool)` (Consolidating `AddSeeds` and `AddNewURLs`):**
    1.  Acquire `mu.Lock()`.
    2.  For each URL in `urls`:
        a.  Normalize the URL (basic normalization like ensuring scheme, removing fragments if desired for uniqueness).
        b.  If `visited[normalizedURL]` is `true`, skip (duplicate).
        c.  Set `visited[normalizedURL] = true`.
        d.  Increment `activeTasks` counter.
        e.  Try to send `normalizedURL` to `urlChan`:
            ```go
            select {
            case m.urlChan <- normalizedURL:
                // Successfully added
            case <-m.shutdownChan:
                // System is shutting down, don't add more
                m.activeTasks-- // Decrement as it won't be processed
                // Unlock before returning if we exit early
            }
            ```
    3.  Release `mu.Unlock()`.

*   **`GetNextURL() (url string, ok bool)`:**
    1.  This method is primarily for Fetcher workers.
    2.  Select on `urlChan` and `shutdownChan`:
        ```go
        select {
        case url, chanOpen := <-m.urlChan:
            if !chanOpen {
                // Channel closed, means no more new URLs will be added AND queue is empty
                return "", false
            }
            return url, true // Got a URL
        case <-m.shutdownChan:
            return "", false // System is shutting down
        }
        ```

*   **`TaskCompleted()`:**
    1.  Called by the system after a URL has been fully processed (fetched, parsed, stored, and new links (if any) added back to URL Manager).
    2.  Acquire `mu.Lock()`.
    3.  Decrement `activeTasks`.
    4.  Check if `activeTasks == 0` AND `len(m.urlChan) == 0`. If so, it means the crawl might be complete.
        *   Set `m.done = true`.
        *   Close `m.urlChan` (to signal fetchers no more URLs will come from this source).
        *   Consider closing `m.shutdownChan` if this is the sole condition for shutdown.
    5.  Release `mu.Unlock()`.

*   **`SignalShutdown()`:**
    1.  Acquire `mu.Lock()`.
    2.  Close `m.shutdownChan` (if not already closed).
    3.  Set `m.done = true`.
    4.  Close `m.urlChan` (if not already closed).
    5.  Release `mu.Unlock()`.
    *   This can be called on Ctrl+C or other termination signals.

*   **`IsDone()`:**
    1.  Acquire `mu.Lock()`.
    2.  `isDone := m.done`
    3.  Release `mu.Unlock()`.
    4.  Return `isDone`.

## 6. Error Handling

*   The URL Manager itself has limited error states for MVP. The primary concern is managing the flow and duplicates.
*   Malformed URLs should ideally be caught before being passed to `AddURLs`, but basic validation/normalization is good practice.

## 7. Concurrency Considerations

*   **`urlChan`:** Go channels are inherently concurrency-safe for sending and receiving.
*   **`visited` map, `activeTasks`, `done` flag:** Access to these shared resources must be protected by `sync.Mutex`.
*   **Graceful Shutdown:** Using `shutdownChan` allows goroutines reading from `urlChan` or performing operations to terminate gracefully when a shutdown is initiated. Closing `urlChan` signals to consumers that no more items will be produced.

## 8. Dependencies

*   Go standard library: `sync`, `net/url` (for normalization).
*   Internal: CLI (for seed URLs), Parser (for new URLs), Fetcher (consumes URLs).

## 9. Design Choices & Reasoning

*   **Channel for Queue (`urlChan`):** Simplifies concurrent access for Fetchers. Buffered channel helps decouple producers (Parsers, initial seeding) from consumers (Fetchers) to a degree.
*   **Mutex for `visited` map and `activeTasks`:** Standard practice for protecting shared mutable data in Go.
*   **`activeTasks` counter:** Essential for determining when all initiated work is truly complete, especially when the `urlChan` might be temporarily empty but tasks are still being processed.
*   **In-Memory for MVP:** Keeps the design simple. For Phase 2/3, this would be replaced by a persistent and distributed queue (e.g., SQS, Redis list, Kafka).
*   **Basic Normalization:** Important for effective duplicate detection.

This LLD for the URL Manager provides a robust in-memory solution for the MVP, focusing on correct concurrency and duplicate handling.
