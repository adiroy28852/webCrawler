# Low-Level Design (LLD): Scheduler / URL Frontier Component

## 1. Introduction

*   **Purpose:** Describe the responsibility of the Scheduler / URL Frontier component in the web crawler.
*   **Scope:** What this component does and does not do. This component is central to managing the crawl process.

## 2. Responsibilities

*   Manages a queue (or multiple queues) of URLs to be crawled (the "URL Frontier").
*   Receives new URLs from the Parser component and initial seed URLs.
*   Provides URLs to Fetcher workers when they are ready for new tasks.
*   Ensures politeness by managing crawl delays for specific domains (e.g., only one request to `example.com` every N seconds).
*   Handles URL prioritization (e.g., crawl high-priority sites first, or based on depth).
*   Detects and handles duplicate URLs to avoid redundant crawling.
*   Potentially manages crawl depth limits.
*   Coordinates the overall crawling process, deciding when to stop or pause.
*   Tracks the state of URLs (e.g., pending, in-progress, crawled, failed).

## 3. Interfaces

*   **Input:**
    *   Seed URLs (initialization).
    *   New URLs from Parser (`ParsedData.new_urls`).
    *   Feedback from Fetcher (e.g., successful crawl, failed attempt, for updating domain-specific politeness timers).
*   **Output:**
    *   `CrawlTask` (or URL string) to Fetcher workers.
*   **Functions/Methods (Conceptual):**
    *   `AddURLs(urls []string)`: Adds a list of new URLs to the frontier.
    *   `GetNextURL() (CrawlTask, error)`: Retrieves the next URL to be crawled, respecting politeness and priority.
    *   `MarkCrawled(url string, success bool)`: Updates the status of a URL.
    *   `IsEmpty() bool`: Checks if the frontier has any URLs left.
    *   `CanFetch(domain string) bool`: Checks if a domain can be fetched based on politeness rules.
    *   `RecordFetchAttempt(domain string)`: Records that a fetch attempt is being made to a domain (for politeness).

## 4. Data Structures

*   **URL Queue(s):** The core data structure. This could be a simple FIFO queue, a priority queue, or multiple queues (e.g., one per domain for politeness, or per priority level).
    *   Consider thread-safety if accessed by multiple goroutines.
*   **Visited Set/Bloom Filter:** To keep track of URLs already added to the queue or crawled, for duplicate detection. A Bloom filter can be memory-efficient for very large sets.
*   **Domain Timestamps:** A map or hash table to store the last crawl time for each domain to enforce politeness delays (e.g., `map[string]time.Time`).
*   `SchedulerConfig`: Struct for configurations like default crawl delay, max depth, priority rules.
*   `URLInfo`: Struct to store URL, priority, depth, status, etc., within the queue.

## 5. Core Logic / Algorithms

*   **URL Addition Logic:**
    1.  Receive new URLs.
    2.  For each URL: Normalize it.
    3.  Check against visited set/Bloom filter. If already seen, discard.
    4.  If new, add to visited set and then to the appropriate queue based on priority/domain.
*   **URL Retrieval Logic (`GetNextURL`):**
    1.  Iterate through queues (if multiple) or check the main queue.
    2.  For a candidate URL, extract its domain.
    3.  Check `DomainTimestamps` to see if the politeness delay for that domain has passed.
    4.  If delay has passed (or no entry for domain yet), return the URL and update `DomainTimestamps` for that domain.
    5.  If delay has not passed, try the next URL/queue or wait.
    6.  Handle empty queue scenario.
*   **Politeness Enforcement:** Detail how domain-specific delays are calculated and enforced.
*   **Duplicate Detection:** Explain the mechanism (e.g., hash set of normalized URLs, Bloom filter).
*   **Priority Handling:** If implementing priority, explain how URLs are selected based on it.

## 6. Error Handling

*   Handling an empty frontier.
*   Errors in adding URLs (e.g., malformed URLs that slip through normalization).
*   Concurrency-related errors if data structures are not properly synchronized.

## 7. Concurrency Considerations

*   The URL Frontier is a critical shared resource if multiple Fetcher and Parser workers are interacting with it.
*   All access to shared data structures (queues, visited set, domain timestamps) must be synchronized using Go's concurrency primitives (e.g., mutexes, channels).
*   Consider using channels to pass URLs to/from the Scheduler to decouple components.

## 8. Dependencies

*   Go standard library: `sync` (for mutexes, waitgroups), `time`, `container/heap` (for priority queue).
*   Other components: Parser (for input of new URLs), Fetcher (for output of tasks and input of crawl status).

*(This is a template. You will fill in the details based on your project decisions.)*
