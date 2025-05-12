# Low-Level Design (LLD): Phase 1 - Parser Component

## 1. Introduction

*   **Purpose:** This document details the Low-Level Design for the Parser component of the Phase 1 MVP web crawler.
*   **Scope:** Defines how fetched HTML content is processed to extract new hyperlinks and relevant page data (specifically, the page title for the MVP).
*   **Reference HLD:** `hld_phase1_mvp.md`
*   **Reference LLD (Fetcher):** `lld_phase1_fetcher.md` (provides input to Parser)

## 2. Responsibilities

*   Receive fetched page data (`FetchedPageData` containing URL and HTML body as `[]byte`) from a Fetcher worker (via a channel).
*   If the fetched page data indicates an error from the Fetcher, bypass parsing and potentially log or signal the error.
*   Parse the HTML content.
*   Extract all hyperlink (`<a>` tag `href` attributes).
*   Normalize extracted URLs (resolve relative URLs to absolute ones using the base URL of the fetched page, ensure scheme presence).
*   Extract the page title (content of the `<title>` tag).
*   Pass the list of new, normalized URLs to the URL Manager (via a channel).
*   Pass the extracted page data (original URL, title) to the Storage Adapter (via a channel).

## 3. Interfaces

*   **Input:**
    *   Receives `FetchedPageData` from a channel connected to Fetcher workers.
    ```go
    // From common package or fetcher LLD
    type FetchedPageData struct {
        URL     string
        Body    []byte
        Error   error
    }
    ```
*   **Output:**
    *   Sends new, normalized URLs (`[]string`) to a channel leading to the URL Manager.
    *   Sends `PageStorageData` (a struct containing original URL and extracted title) to a channel leading to the Storage Adapter.
    ```go
    // Package: common (or parser if specific)
    type PageStorageData struct {
        URL   string
        Title string
        // Error error // Optional: if parser itself has distinct reportable errors for storage
    }
    ```
*   **Functions/Methods (Conceptual, for a Parser worker goroutine):**
    *   `RunWorker(id int, fetchedDataChan <-chan FetchedPageData, newLinksChan chan<- []string, pageStoreChan chan<- PageStorageData, shutdownChan <-chan struct{}, wg *sync.WaitGroup)`: The main loop for a parser worker.

## 4. Data Structures

*   The `PageStorageData` struct defined above for output to storage.
*   No complex internal state per Parser worker beyond what's needed for a single parse operation.

## 5. Core Logic / Algorithms (within a Parser worker goroutine)

1.  **Worker Loop:**
    *   The worker runs in a loop, listening on `fetchedDataChan` and `shutdownChan`.
    *   `for { select { case data, ok := <-fetchedDataChan: ... case <-shutdownChan: return } }`
2.  **Receive Fetched Data:**
    *   If `fetchedDataChan` is closed (`!ok`), the worker terminates.
    *   If `data.Error != nil` from `FetchedPageData`, log this error (e.g., "Skipping parsing for URL [URL] due to fetch error: [Error]"). Do not proceed with parsing for this item. Signal task completion for this URL to the URL Manager (important for `activeTasks` counter).
3.  **Parse HTML:**
    *   Convert `data.Body` (`[]byte`) to an `io.Reader` (e.g., `bytes.NewReader(data.Body)`).
    *   Parse the HTML using `golang.org/x/net/html.Parse(reader)`.
    *   If `html.Parse` returns an error, log it (e.g., "Failed to parse HTML for URL [URL]: [Error]"). Do not proceed further for this item. Signal task completion.
4.  **Extract Links and Title (Recursive Traversal Function):**
    *   Define a recursive function, say `extract(node *html.Node, baseURL *url.URL, links *[]string, title *string)`.
    *   **Base Case:** If `node == nil`, return.
    *   **Title Extraction:** If `node.Type == html.ElementNode && node.Data == "title"`:
        *   If `node.FirstChild != nil && node.FirstChild.Type == html.TextNode`, set `*title = node.FirstChild.Data` (take the first title found).
    *   **Link Extraction:** If `node.Type == html.ElementNode && node.Data == "a"`:
        *   Iterate through `node.Attr`.
        *   If `attr.Key == "href"`, get `attr.Val` (the link URL string).
            *   Parse this link string: `parsedLink, err := url.Parse(attr.Val)`.
            *   If `err != nil`, skip this malformed link.
            *   Resolve relative links: `absoluteLink := baseURL.ResolveReference(parsedLink)`.
            *   Convert `absoluteLink` back to string and add to `*links` slice.
    *   **Recursive Step:** Call `extract` for `node.FirstChild` and then for `node.NextSibling`.
5.  **Initiate Extraction:**
    *   `parsedBaseURL, err := url.Parse(data.URL)` (handle error if base URL is invalid, though it should be valid if fetched).
    *   `var extractedLinks []string`
    *   `var pageTitle string`
    *   Call `extract(docNode, parsedBaseURL, &extractedLinks, &pageTitle)` starting with the root `docNode` from `html.Parse`.
6.  **Process and Send Extracted Data:**
    *   **Send New Links:** If `len(extractedLinks) > 0`, send `extractedLinks` to `newLinksChan` (for the URL Manager).
    *   **Send Page Data for Storage:** Create `PageStorageData{URL: data.URL, Title: pageTitle}` and send to `pageStoreChan`.
7.  After sending data (or determining there's nothing to send/error occurred), the worker signals task completion for this URL to the URL Manager.
8.  The worker then loops back to await the next `FetchedPageData` or shutdown signal.

## 6. Error Handling

*   **Fetcher Errors:** Handled by not attempting to parse if `FetchedPageData.Error` is set.
*   **HTML Parsing Errors:** Logged. The specific page processing is aborted.
*   **URL Parsing Errors (for extracted links):** Malformed `href` values are skipped.
*   **Base URL Parsing Error:** Should be rare if the URL was successfully fetched. If it occurs, link resolution will fail; log and potentially skip link extraction.

## 7. Concurrency Considerations

*   Each Parser worker runs as an independent goroutine.
*   They are stateless concerning individual parse operations.
*   They consume data from a shared channel (`fetchedDataChan`) and send results to other shared channels (`newLinksChan`, `pageStoreChan`).
*   The `golang.org/x/net/html` parsing functions are safe for concurrent use as they operate on their own input data.

## 8. Dependencies

*   Go standard library: `golang.org/x/net/html`, `net/url`, `io`, `bytes`, `strings`, `sync`, `fmt`.
*   Internal: Fetcher (for input `FetchedPageData`), URL Manager (for outputting new links), Storage Adapter (for outputting page data to store).

## 9. Design Choices & Reasoning

*   **Worker Pool Model:** Allows concurrent parsing, which can be CPU-bound, improving throughput.
*   **Channel-based Communication:** Decouples Parsers from Fetchers and downstream components.
*   **Standard `golang.org/x/net/html`:** A robust and standard library for HTML parsing in Go. Avoids external dependencies for this core task in MVP.
*   **Recursive Traversal for Extraction:** A common and effective way to walk the HTML DOM tree.
*   **Simple Title Extraction:** For MVP, just grabbing the text content of the first `<title>` tag is sufficient.
*   **Error Handling Strategy:** Prioritize logging errors and continuing if possible, rather than crashing the worker or application for individual page parsing issues.

This LLD for the Parser component outlines a clear path for implementing HTML content processing for the MVP.
