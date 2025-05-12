# Low-Level Design (LLD): Phase 1 - CLI Interface

## 1. Introduction

*   **Purpose:** This document details the Low-Level Design for the Command Line Interface (CLI) component of the Phase 1 MVP web crawler.
*   **Scope:** Defines how users interact with the crawler via the command line, including available commands, arguments, and expected output for basic operations.
*   **Reference HLD:** `hld_phase1_mvp.md`

## 2. Responsibilities

*   Parse command-line arguments and flags provided by the user.
*   Validate user input for basic correctness (e.g., presence of seed URLs).
*   Initiate the crawling process based on parsed arguments.
*   Display basic status information during the crawl (e.g., number of URLs in queue, pages crawled).
*   Provide a way to gracefully stop the crawler (e.g., via Ctrl+C, though full graceful shutdown might be simplified for MVP).
*   Output final summary or errors upon completion or termination.

## 3. Interfaces

*   **Input (Command Line):**
    *   The application will be invoked as a single executable (e.g., `./gocrawler_mvp`).
    *   **Arguments/Flags (using Go `flag` package):**
        *   `-seeds`: (Required) A comma-separated string of initial URLs to start crawling (e.g., `-seeds "http://example.com,http://anotherexample.org"`).
        *   `-workers`: (Optional) Number of concurrent fetcher/parser worker pairs. Defaults to a sensible value (e.g., 3). Example: `-workers 5`.
        *   `-delay`: (Optional) A simple global delay in milliseconds between fetch requests (for basic politeness). Defaults to e.g., 1000ms. Example: `-delay 500`.
        *   `-dbconn`: (Optional) PostgreSQL connection string. Defaults to a common local setup (e.g., `"postgres://user:password@localhost/crawlerdb?sslmode=disable"`). Example: `-dbconn "postgres://myuser:mypass@host:port/mydb?sslmode=verify-full"`.
        *   `-useragent`: (Optional) Custom User-Agent string for HTTP requests. Defaults to a generic Go crawler agent. Example: `-useragent "MyLearningCrawler/1.0"`.
*   **Output (Standard Output/Error):**
    *   Startup messages (e.g., "Crawler starting with N workers...").
    *   Periodic status updates (e.g., "Queue: X, Crawled: Y, Errors: Z").
    *   Error messages (e.g., invalid arguments, database connection failure).
    *   Completion message (e.g., "Crawl finished. Total pages crawled: A").
*   **Internal Interaction:**
    *   Passes parsed configuration (seeds, worker count, etc.) to the Configuration Manager or directly to the main application controller.
    *   Triggers the main crawling logic.

## 4. Data Structures

*   A struct to hold the parsed command-line arguments:
    ```go
    type CLIFlags struct {
        SeedURLs        []string // Parsed from comma-separated string
        NumWorkers      int
        CrawlDelay      time.Duration // Parsed from int milliseconds
        DBConnectionString string
        UserAgent       string
    }
    ```

## 5. Core Logic / Algorithms

*   **Argument Parsing (`main` function):**
    1.  Define flags using the `flag` package (e.g., `flag.String("seeds", "", "Comma-separated seed URLs")`).
    2.  Call `flag.Parse()`.
    3.  Retrieve flag values.
*   **Input Validation:**
    1.  Check if `-seeds` flag is provided. If not, print usage and exit.
    2.  Parse the comma-separated seed URL string into a `[]string`.
    3.  Validate basic format of seed URLs (e.g., using `net/url.Parse`).
    4.  Validate `NumWorkers` (e.g., must be > 0).
    5.  Validate `CrawlDelay` (e.g., must be >= 0).
*   **Status Display:**
    *   A simple loop in a separate goroutine (or integrated into the main controller) that periodically prints statistics. For MVP, this can be very basic, perhaps just printing a line every few seconds.
*   **Initiating Crawl:**
    *   After parsing and validation, the CLI component will call a function in the main application controller/scheduler, passing the validated configuration.

## 6. Error Handling

*   **Invalid Arguments:** If required arguments are missing or formats are incorrect, print a helpful usage message to standard error and exit with a non-zero status code.
*   **URL Parsing Errors:** If seed URLs are malformed, report and exit.

## 7. Concurrency Considerations

*   The primary CLI logic (argument parsing) runs in the main goroutine.
*   Status display might run in a separate goroutine if it needs to be asynchronous to the main crawl loop.
*   The CLI will trigger the main crawling process, which will involve significant concurrency (worker pools), but the CLI component itself is mostly concerned with setup and initial triggering.

## 8. Dependencies

*   Go standard library: `flag`, `fmt`, `os`, `strings`, `net/url`, `time`.

## 9. Example Invocation

```bash
./gocrawler_mvp -seeds "http://books.toscrape.com" -workers 4 -delay 1000
```

This LLD provides a clear plan for the CLI component, focusing on simplicity for the MVP while ensuring essential functionalities are covered.
