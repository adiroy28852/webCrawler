# Low-Level Design (LLD): Fetcher Component

## 1. Introduction

*   **Purpose:** Describe the responsibility of the Fetcher component in the web crawler.
*   **Scope:** What this component does and does not do.

## 2. Responsibilities

*   Receives a URL to fetch from the Scheduler/URL Frontier.
*   Makes an HTTP GET request to the given URL.
*   Handles HTTP response codes (e.g., 200 OK, 3xx redirects, 4xx client errors, 5xx server errors).
*   Implements retry mechanisms for transient errors (e.g., network issues, temporary server errors).
*   Respects `robots.txt` rules (may coordinate with a separate Robots.txt handler or have its own logic).
*   Manages request headers (e.g., User-Agent).
*   Handles timeouts for requests.
*   Returns the fetched content (e.g., HTML body) and relevant metadata (e.g., response headers, status code, final URL after redirects) to the Parser.

## 3. Interfaces

*   **Input:**
    *   `CrawlTask` (or similar struct): Contains the URL to fetch, potentially priority, depth, etc.
*   **Output:**
    *   `FetchedPage` (or similar struct): Contains the raw page content (e.g., `[]byte` or `string`), the final URL (after redirects), HTTP status code, response headers, and any error encountered.
*   **Functions/Methods (Conceptual):**
    *   `Fetch(task CrawlTask) (FetchedPage, error)`: Main function to perform the fetching operation.

## 4. Data Structures

*   `FetcherConfig`: Struct to hold configuration like User-Agent, request timeout, max retries, delay between retries.
*   Internal data structures for managing retries, cookies (if needed), etc.

## 5. Core Logic / Algorithms

*   **Request Execution Flow:**
    1.  Receive `CrawlTask`.
    2.  Check `robots.txt` compliance (if applicable within this component).
    3.  Construct HTTP request.
    4.  Execute request using an HTTP client.
    5.  Handle response:
        *   Success (2xx): Read body, prepare `FetchedPage`.
        *   Redirect (3xx): Update URL, potentially re-queue or follow (up to a limit).
        *   Client Error (4xx): Log error, mark URL as failed (e.g., 404 Not Found).
        *   Server Error (5xx): Implement retry logic based on `FetcherConfig`.
    6.  Return `FetchedPage` or error.
*   **Retry Logic:** Detail the strategy (e.g., exponential backoff).
*   **Timeout Handling:** How timeouts are detected and handled.

## 6. Error Handling

*   List potential errors (network errors, HTTP errors, parsing errors if any basic validation is done here).
*   How each error type is handled and reported.

## 7. Concurrency Considerations (if applicable within Fetcher itself)

*   If the Fetcher manages its own pool of goroutines for making requests (less common if workers are managed by a higher-level Scheduler).

## 8. Dependencies

*   Go standard library: `net/http`, `time`.
*   External libraries (if any, e.g., more advanced HTTP clients).
*   Other components: Scheduler/URL Frontier (for input), Parser (for output).

*(This is a template. You will fill in the details based on your project decisions.)*
