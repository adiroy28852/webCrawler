# Low-Level Design (LLD): Phase 1 - Fetcher Component

## 1. Introduction

*   **Purpose:** This document details the Low-Level Design for the Fetcher component of the Phase 1 MVP web crawler.
*   **Scope:** Defines how web pages are retrieved from URLs provided by the URL Manager. It covers HTTP requests, basic error handling, and retry mechanisms suitable for an MVP.
*   **Reference HLD:** `hld_phase1_mvp.md`
*   **Reference LLD (URL Manager):** `lld_phase1_url_manager.md`
*   **Reference LLD (Config Manager):** `lld_phase1_config_manager.md`

## 2. Responsibilities

*   Receive a URL to fetch from the URL Manager (via a channel).
*   Construct and execute an HTTP GET request to the given URL.
*   Set appropriate HTTP headers, including a User-Agent string obtained from the Configuration Manager.
*   Handle HTTP response codes (e.g., 200 OK, 3xx redirects, 4xx client errors, 5xx server errors).
*   Implement a simple retry mechanism for transient errors (e.g., network issues, temporary server errors) based on configuration.
*   Adhere to a simple global crawl delay (obtained from Configuration Manager) before making a request.
*   Handle request timeouts.
*   Pass the fetched HTML content (as `[]byte`) and the final URL (after any redirects) to a Parser worker (via a channel).
*   Report errors if a page cannot be fetched successfully after retries.

## 3. Interfaces

*   **Input:**
    *   Receives URLs (`string`) from the `urlChan` provided by the URL Manager.
    *   Accesses configuration settings from the Configuration Manager (e.g., User-Agent, crawl delay, retry count, timeout).
*   **Output:**
    *   Sends `FetchedPageData` (a struct containing URL, HTML content as `[]byte`, and potentially error status) to a channel leading to Parser workers.
    ```go
    // Package: common (or fetcher if specific)
    type FetchedPageData struct {
        URL     string
        Body    []byte
        Error   error // nil if successful, otherwise contains fetch error
    }
    ```
*   **Functions/Methods (Conceptual, for a Fetcher worker goroutine):**
    *   `RunWorker(id int, urlSource <-chan string, resultChan chan<- FetchedPageData, config *config.Manager, shutdownChan <-chan struct{}, wg *sync.WaitGroup)`: The main loop for a fetcher worker.

## 4. Data Structures

*   No complex internal state per Fetcher worker beyond what's needed for a single fetch operation (e.g., HTTP client instance).
*   The `FetchedPageData` struct defined above is key for output.

## 5. Core Logic / Algorithms (within a Fetcher worker goroutine)

1.  **Worker Loop:**
    *   The worker runs in a loop, listening on `urlSource` channel and `shutdownChan`.
    *   `for { select { case url, ok := <-urlSource: ... case <-shutdownChan: return } }`
2.  **Receive URL:**
    *   If `urlSource` channel is closed (`!ok`), the worker terminates.
3.  **Apply Global Crawl Delay:**
    *   `time.Sleep(config.GetCrawlDelay())` before making the request.
4.  **Fetch Attempt Loop (for retries):**
    *   `maxRetries := config.GetMaxRetries()` (e.g., default to 2-3 for MVP).
    *   `for attempt := 0; attempt < maxRetries; attempt++ { ... }`
5.  **Construct HTTP Request:**
    *   `req, err := http.NewRequest("GET", url, nil)`
    *   Set User-Agent: `req.Header.Set("User-Agent", config.GetUserAgent())`
    *   (For MVP, we won't handle `robots.txt` in the Fetcher directly. A global delay is the only politeness.)
6.  **Execute HTTP Request:**
    *   Create an `http.Client` with a timeout: `client := http.Client{Timeout: config.GetRequestTimeout()}` (e.g., 10-15 seconds).
    *   `resp, err := client.Do(req)`
    *   **Error Handling (Network/Timeout):** If `err != nil` (e.g., DNS error, connection refused, timeout):
        *   Log the error.
        *   If `attempt < maxRetries-1`, sleep for a short duration (e.g., `config.GetRetryDelay()`) and `continue` to the next attempt.
        *   If last attempt, prepare `FetchedPageData{URL: url, Error: err}` and send to `resultChan`. Break from retry loop.
7.  **Handle HTTP Response:**
    *   `defer resp.Body.Close()`
    *   **Success (2xx):**
        *   `if resp.StatusCode >= 200 && resp.StatusCode < 300:`
        *   Read response body: `body, readErr := io.ReadAll(resp.Body)`.
        *   If `readErr != nil`, prepare `FetchedPageData{URL: url, Error: readErr}`.
        *   Else, prepare `FetchedPageData{URL: resp.Request.URL.String(), Body: body, Error: nil}` (Use `resp.Request.URL` to get the final URL after redirects).
        *   Send to `resultChan`. Break from retry loop.
    *   **Redirects (3xx):** For MVP, the standard `http.Client` follows redirects by default. The final URL is available in `resp.Request.URL`. If a non-redirecting client were used, manual redirect handling would be needed here (out of scope for MVP LLD).
    *   **Client Errors (4xx, e.g., 404 Not Found, 403 Forbidden):**
        *   Log the error (e.g., "URL not found: [URL], Status: [Code]").
        *   Prepare `FetchedPageData{URL: url, Error: fmt.Errorf("HTTP Error: %d", resp.StatusCode)}`.
        *   Send to `resultChan`. Break from retry loop (no point retrying 4xx errors).
    *   **Server Errors (5xx):**
        *   Log the error.
        *   If `attempt < maxRetries-1`, sleep and `continue` to retry.
        *   If last attempt, prepare `FetchedPageData{URL: url, Error: fmt.Errorf("HTTP Error: %d", resp.StatusCode)}`.
        *   Send to `resultChan`. Break from retry loop.
8.  **Send Result:** The `FetchedPageData` (with body or error) is sent to `resultChan`.
9.  The worker then loops back to await the next URL or shutdown signal.

## 6. Error Handling (Summary)

*   **Network Errors/Timeouts:** Retried up to `maxRetries`.
*   **HTTP 4xx Errors:** Not retried. Reported as an error for the URL.
*   **HTTP 5xx Errors:** Retried up to `maxRetries`.
*   **Body Read Errors:** Reported as an error for the URL.
*   All errors are packaged into the `FetchedPageData.Error` field.

## 7. Concurrency Considerations

*   Each Fetcher worker runs as an independent goroutine.
*   They are stateless concerning individual fetch operations.
*   They consume URLs from a shared channel (`urlSource`) and send results to another shared channel (`resultChan`).
*   The `http.Client` is safe for concurrent use by multiple goroutines if its Transport is not modified after creation (which is the case here as we create a new client per request or per worker, or use the default client which is also safe).

## 8. Dependencies

*   Go standard library: `net/http`, `io`, `fmt`, `time`, `sync`.
*   Internal: Configuration Manager (for settings), URL Manager (for `urlSource`), Parser (for `resultChan` destination).

## 9. Design Choices & Reasoning

*   **Worker Pool Model:** Allows concurrent fetching, which is essential for performance.
*   **Channel-based Communication:** Decouples Fetchers from the URL Manager and Parsers, adhering to Go concurrency patterns.
*   **Simple Global Delay for MVP Politeness:** Easy to implement. More sophisticated per-domain politeness is deferred to Phase 2.
*   **Configurable Retries and Timeouts:** Provides basic robustness.
*   **Standard `net/http` Client:** Sufficient for MVP needs. It handles redirects automatically by default.
*   **Error Propagation:** Errors are explicitly passed along in `FetchedPageData` for downstream components (e.g., Parser or main controller) to decide how to handle or log them.

This LLD for the Fetcher component provides a clear path for implementing the page retrieval logic for the MVP.
