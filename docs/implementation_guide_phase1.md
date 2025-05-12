# Implementation Guide: Phase 1 - MVP CLI-based Web Crawler

## 0. Introduction

This guide provides step-by-step instructions to build the Phase 1 MVP of our web crawler. It assumes you have reviewed and understood the following design documents:

*   **High-Level Design:** `hld_phase1_mvp.md`
*   **Low-Level Designs (LLDs):**
    *   `lld_phase1_cli.md`
    *   `lld_phase1_config_manager.md`
    *   `lld_phase1_url_manager.md`
    *   `lld_phase1_fetcher.md`
    *   `lld_phase1_parser.md`
    *   `lld_phase1_storage_postgres.md`

We will build the crawler component by component, following these LLDs, and then integrate them into a working CLI application.

**Prerequisites:**
*   Go (version 1.18+ recommended) installed and configured.
*   PostgreSQL server installed and running.
*   Basic understanding of Go syntax and concepts.
*   A text editor or IDE for Go development.

## 1. Project Setup

Refer to your HLD for the overall project structure. Let's create it:

1.  **Create Project Directory:**
    ```bash
    mkdir gocrawler_mvp
    cd gocrawler_mvp
    ```

2.  **Initialize Go Module:**
    ```bash
    go mod init gocrawler_mvp
    ```
    (You can replace `gocrawler_mvp` with your chosen module name if different, e.g., `github.com/yourusername/gocrawler_mvp`)

3.  **Create Initial Directory Structure (as per HLD/LLDs):**
    ```bash
    mkdir cmd cli config common urlmanager fetcher parser storage
    touch cmd/main.go
    touch cli/cli.go
    touch config/config.go
    touch common/types.go
    touch urlmanager/urlmanager.go
    touch fetcher/fetcher.go
    touch parser/parser.go
    touch storage/storage.go
    ```
    Also, copy your HLD and LLD markdown files into this project directory (e.g., in a `docs/` subfolder or at the root).

## 2. Setting up PostgreSQL Database

Refer to `lld_phase1_storage_postgres.md` for the schema.

1.  **Access PostgreSQL:** Open `psql` or your preferred PostgreSQL client.
2.  **Create Database (if it doesn't exist):**
    ```sql
    CREATE DATABASE crawler_mvp_db;
    ```
    (Adjust name if your default connection string in `lld_phase1_config_manager.md` uses a different DB name).
3.  **Connect to the Database:**
    ```sql
    \c crawler_mvp_db
    ```
4.  **Create `pages` Table:**
    ```sql
    CREATE TABLE IF NOT EXISTS pages (
        id SERIAL PRIMARY KEY,
        url TEXT NOT NULL UNIQUE,
        title TEXT,
        crawled_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    ```
5.  **Verify:** You can use `\dt` to list tables and see `pages`.
6.  **User and Permissions:** Ensure you have a PostgreSQL user with permissions to connect to this database and perform CRUD operations on the `pages` table. The connection string in your `config/config.go` will use these credentials.

## 3. Implementing Core Data Structures (`common/types.go`)

Based on the LLDs, let's define shared data structures. Open `common/types.go`:

```go
// common/types.go
package common

// FetchedPageData holds data retrieved by a Fetcher.
type FetchedPageData struct {
    URL   string
    Body  []byte
    Error error
}

// PageStorageData holds data to be stored by the StorageAdapter.
type PageStorageData struct {
    URL   string
    Title string
}

// Add other common types if identified during LLD refinement.
```

## 4. Implementing the Configuration Manager (`config/config.go`)

Follow `lld_phase1_config_manager.md`.

```go
// config/config.go
package config

import (
    "time"
    // You might need to import your cli package if CLIFlags is defined there
    // For now, let's assume CLIFlags will be passed as simple types from main.
)

const (
    DefaultNumWorkers       = 3
    DefaultCrawlDelay       = 1 * time.Second
    DefaultDBConnectionString = "postgres://postgres:password@localhost:5432/crawler_mvp_db?sslmode=disable" // CHANGE THIS!
    DefaultUserAgent        = "GoCrawlerMVP/0.1"
    DefaultMaxRetries       = 2
    DefaultRequestTimeout   = 15 * time.Second
    DefaultRetryDelay       = 2 * time.Second
)

type Manager struct {
    SeedURLs         []string
    NumWorkers       int
    CrawlDelay       time.Duration
    DBConnectionString string
    UserAgent        string
    MaxRetries       int
    RequestTimeout   time.Duration
    RetryDelay       time.Duration
}

// NewManager creates and initializes a new configuration Manager.
// It takes parsed CLI values as input.
func NewManager(seeds []string, numWorkers int, crawlDelayMs int, dbConn, userAgent string) *Manager {
    cfg := &Manager{
        SeedURLs:         seeds,
        NumWorkers:       DefaultNumWorkers,
        CrawlDelay:       DefaultCrawlDelay,
        DBConnectionString: DefaultDBConnectionString,
        UserAgent:        DefaultUserAgent,
        MaxRetries:       DefaultMaxRetries,
        RequestTimeout:   DefaultRequestTimeout,
        RetryDelay:       DefaultRetryDelay,
    }

    if numWorkers > 0 {
        cfg.NumWorkers = numWorkers
    }
    if crawlDelayMs > 0 {
        cfg.CrawlDelay = time.Duration(crawlDelayMs) * time.Millisecond
    }
    if dbConn != "" {
        cfg.DBConnectionString = dbConn
    }
    if userAgent != "" {
        cfg.UserAgent = userAgent
    }
    // Add similar checks for MaxRetries, RequestTimeout, RetryDelay if they become CLI flags

    return cfg
}

// Getter methods as defined in LLD
func (m *Manager) GetSeedURLs() []string         { return m.SeedURLs }
func (m *Manager) GetNumWorkers() int            { return m.NumWorkers }
func (m *Manager) GetCrawlDelay() time.Duration  { return m.CrawlDelay }
func (m *Manager) GetDBConnectionString() string { return m.DBConnectionString }
func (m *Manager) GetUserAgent() string          { return m.UserAgent }
func (m *Manager) GetMaxRetries() int            { return m.MaxRetries }
func (m *Manager) GetRequestTimeout() time.Duration { return m.RequestTimeout }
func (m *Manager) GetRetryDelay() time.Duration    { return m.RetryDelay }

```
**Note:** Remember to change the `DefaultDBConnectionString` to match your actual PostgreSQL setup if it differs, or ensure the CLI flag is always used.

## 5. Implementing the CLI Interface (`cli/cli.go`)

Follow `lld_phase1_cli.md`.

```go
// cli/cli.go
package cli

import (
    "flag"
    "fmt"
    "os"
    "strings"
    "net/url"
)

// CLIFlags holds the parsed command-line arguments.
// This struct can be moved to common/types.go if shared more widely.
type CLIFlags struct {
    SeedURLsRaw      string // Comma-separated
    SeedURLsParsed   []string
    NumWorkers       int
    CrawlDelayMs     int
    DBConnectionString string
    UserAgent        string
}

func ParseFlags() (*CLIFlags, error) {
    cf := &CLIFlags{}

    flag.StringVar(&cf.SeedURLsRaw, "seeds", "", "(Required) Comma-separated seed URLs (e.g., \"http://example.com,http://anotherexample.org\")")
    flag.IntVar(&cf.NumWorkers, "workers", config.DefaultNumWorkers, "Number of concurrent fetcher/parser worker pairs")
    flag.IntVar(&cf.CrawlDelayMs, "delay", int(config.DefaultCrawlDelay/time.Millisecond), "Global delay in milliseconds between fetch requests")
    flag.StringVar(&cf.DBConnectionString, "dbconn", "", fmt.Sprintf("PostgreSQL connection string (default: \"%s\")", config.DefaultDBConnectionString))
    flag.StringVar(&cf.UserAgent, "useragent", "", fmt.Sprintf("Custom User-Agent string (default: \"%s\")", config.DefaultUserAgent))

    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
        flag.PrintDefaults()
    }

    flag.Parse()

    if cf.SeedURLsRaw == "" {
        flag.Usage()
        return nil, fmt.Errorf("-seeds flag is required")
    }

    rawUrls := strings.Split(cf.SeedURLsRaw, ",")
    for _, rawUrl := range rawUrls {
        trimmedUrl := strings.TrimSpace(rawUrl)
        if trimmedUrl == "" {
            continue
        }
        _, err := url.ParseRequestURI(trimmedUrl) // Basic validation
        if err != nil {
            flag.Usage()
            return nil, fmt.Errorf("invalid seed URL %s: %v", trimmedUrl, err)
        }
        cf.SeedURLsParsed = append(cf.SeedURLsParsed, trimmedUrl)
    }

    if len(cf.SeedURLsParsed) == 0 {
        flag.Usage()
        return nil, fmt.Errorf("-seeds flag must contain at least one valid URL")
    }

    if cf.NumWorkers <= 0 {
        return nil, fmt.Errorf("-workers must be a positive integer")
    }
    if cf.CrawlDelayMs < 0 {
        return nil, fmt.Errorf("-delay must be a non-negative integer")
    }

    return cf, nil
}
```
*You will need to import your `config` package in `cli.go` to access defaults if you structure it this way, or pass defaults into `ParseFlags`.* For simplicity in this guide, `config.DefaultNumWorkers` etc. are used directly, assuming `config` package is accessible.

## 6. Implementing the URL Manager (`urlmanager/urlmanager.go`)

Follow `lld_phase1_url_manager.md`.

```go
// urlmanager/urlmanager.go
package urlmanager

import (
    "sync"
    "net/url"
    "strings"
    "gocrawler_mvp/common" // Assuming common types are here
)

const urlChannelBuffer = 100 // Example buffer size

type Manager struct {
    mu             sync.Mutex
    visited        map[string]bool
    urlChan        chan string
    shutdownChan   chan struct{}
    activeTasks    int
    isDone         bool
    wg             sync.WaitGroup // To wait for tasks to complete before closing urlChan
}

func NewManager() *Manager {
    return &Manager{
        visited:      make(map[string]bool),
        urlChan:      make(chan string, urlChannelBuffer),
        shutdownChan: make(chan struct{}),
    }
}

// normalizeURL basic normalization for MVP
func normalizeURL(inputURL string, base *url.URL) (string, error) {
    parsed, err := url.Parse(strings.TrimSpace(inputURL))
    if err != nil {
        return "", err
    }
    if base != nil {
        parsed = base.ResolveReference(parsed)
    }
    if parsed.Scheme == "" {
        // Assume http if scheme is missing, or handle as error
        // For now, let's require a scheme from resolved URLs
        if base == nil || base.Scheme == "" {
             return "", fmt.Errorf("cannot determine scheme for URL: %s", parsed.String())
        }
    }
    parsed.Fragment = "" // Remove fragments for uniqueness
    return parsed.String(), nil
}

func (m *Manager) AddURLs(urls []string, baseURL *url.URL) {
    m.mu.Lock()
    defer m.mu.Unlock()

    if m.isDone { // Don't add if already shutting down/done
        return
    }

    for _, u := range urls {
        normalizedU, err := normalizeURL(u, baseURL)
        if err != nil {
            fmt.Printf("[URLManager] Error normalizing URL %s (base %v): %v\n", u, baseURL, err)
            continue
        }

        if !m.visited[normalizedU] {
            m.visited[normalizedU] = true
            m.activeTasks++
            m.wg.Add(1) // Increment WaitGroup counter for each task added

            select {
            case m.urlChan <- normalizedU:
                // fmt.Printf("[URLManager] Added to queue: %s\n", normalizedU)
            case <-m.shutdownChan:
                fmt.Println("[URLManager] Shutdown signaled, not adding more URLs.")
                m.activeTasks-- // Task won't be processed
                m.wg.Done()     // Decrement WaitGroup as it won't run
                return // Exit if shutdown
            }
        }
    }
}

func (m *Manager) GetNextURL() (string, bool) {
    select {
    case u, ok := <-m.urlChan:
        if !ok { // Channel closed
            return "", false
        }
        return u, true
    case <-m.shutdownChan:
        return "", false
    }
}

func (m *Manager) TaskCompleted() {
    m.mu.Lock()
    m.activeTasks--
    // fmt.Printf("[URLManager] Task completed. Active tasks: %d\n", m.activeTasks)
    m.mu.Unlock()
    m.wg.Done() // Decrement WaitGroup counter
}

// CheckDone waits for all tasks to complete and then closes urlChan.
// This should be run in a separate goroutine.
func (m *Manager) CheckDone() {
    m.wg.Wait() // Wait for all tasks (wg.Add/Done calls) to complete

    m.mu.Lock()
    if !m.isDone { // Ensure this runs only once
        fmt.Println("[URLManager] All tasks completed. Closing URL channel.")
        close(m.urlChan)
        m.isDone = true
    }
    m.mu.Unlock()
}

func (m *Manager) SignalShutdown() {
    m.mu.Lock()
    if !m.isDone { // Check if already shut down
        fmt.Println("[URLManager] Shutdown signal received.")
        close(m.shutdownChan) // Signal all listeners
        m.isDone = true       // Mark as done to prevent adding new URLs
        // We don't close urlChan here immediately; CheckDone will do it after wg.Wait()
        // or if GetNextURL sees shutdownChan closed, it will also stop fetchers.
    }
    m.mu.Unlock()
}

func (m *Manager) IsDone() bool {
    m.mu.Lock()
    defer m.mu.Unlock()
    return m.isDone && m.activeTasks == 0 && len(m.urlChan) == 0
}

```

## 7. Implementing the Fetcher (`fetcher/fetcher.go`)

Follow `lld_phase1_fetcher.md`.

```go
// fetcher/fetcher.go
package fetcher

import (
    "fmt"
    "io"
    "net/http"
    "sync"
    "time"

    "gocrawler_mvp/common"
    "gocrawler_mvp/config"
)

// RunWorker is the main loop for a fetcher worker goroutine.
func RunWorker(
    id int,
    cfg *config.Manager,
    urlSource <-chan string, // From URLManager
    resultChan chan<- common.FetchedPageData, // To Parser
    shutdownChan <-chan struct{}, // From URLManager or main
    wg *sync.WaitGroup,
) {
    defer wg.Done()
    fmt.Printf("[Fetcher %d] Starting\n", id)

    client := &http.Client{
        Timeout: cfg.GetRequestTimeout(),
    }

    for {
        select {
        case urlToFetch, ok := <-urlSource:
            if !ok {
                fmt.Printf("[Fetcher %d] URL source channel closed. Exiting.\n", id)
                return
            }

            fmt.Printf("[Fetcher %d] Fetching: %s\n", id, urlToFetch)
            time.Sleep(cfg.GetCrawlDelay()) // Global crawl delay

            var fetchedData common.FetchedPageData
            fetchedData.URL = urlToFetch

            for attempt := 0; attempt < cfg.GetMaxRetries(); attempt++ {
                req, err := http.NewRequest("GET", urlToFetch, nil)
                if err != nil {
                    fetchedData.Error = fmt.Errorf("failed to create request: %v", err)
                    // Non-retryable for this specific error type
                    break 
                }
                req.Header.Set("User-Agent", cfg.GetUserAgent())

                resp, err := client.Do(req)
                if err != nil {
                    fetchedData.Error = fmt.Errorf("request failed (attempt %d/%d): %v", attempt+1, cfg.GetMaxRetries(), err)
                    if attempt < cfg.GetMaxRetries()-1 {
                        time.Sleep(cfg.GetRetryDelay())
                        continue // Retry
                    }
                    break // Max retries reached
                }

                // Use resp.Request.URL.String() as it might have changed after redirects
                fetchedData.URL = resp.Request.URL.String()

                if resp.StatusCode >= 200 && resp.StatusCode < 300 {
                    bodyBytes, readErr := io.ReadAll(resp.Body)
                    resp.Body.Close()
                    if readErr != nil {
                        fetchedData.Error = fmt.Errorf("failed to read body: %v", readErr)
                    } else {
                        fetchedData.Body = bodyBytes
                        fetchedData.Error = nil // Success
                    }
                    break // Success, exit retry loop
                } else if resp.StatusCode >= 400 && resp.StatusCode < 500 {
                    fetchedData.Error = fmt.Errorf("client error HTTP %d for %s", resp.StatusCode, urlToFetch)
                    resp.Body.Close()
                    break // Don't retry 4xx errors
                } else if resp.StatusCode >= 500 {
                    fetchedData.Error = fmt.Errorf("server error HTTP %d for %s (attempt %d/%d)", resp.StatusCode, urlToFetch, attempt+1, cfg.GetMaxRetries())
                    resp.Body.Close()
                    if attempt < cfg.GetMaxRetries()-1 {
                        time.Sleep(cfg.GetRetryDelay())
                        continue // Retry 5xx errors
                    }
                    break // Max retries for 5xx
                } else {
                     fetchedData.Error = fmt.Errorf("unexpected HTTP status %d for %s", resp.StatusCode, urlToFetch)
                     resp.Body.Close()
                     break // Unexpected, don't retry by default
                }
            } // End retry loop
            
            // Send result to parser, regardless of error or success
            select {
            case resultChan <- fetchedData:
            case <-shutdownChan:
                fmt.Printf("[Fetcher %d] Shutdown signaled while sending to parser. Exiting.\n", id)
                return
            }

        case <-shutdownChan:
            fmt.Printf("[Fetcher %d] Shutdown signal received. Exiting.\n", id)
            return
        }
    }
}

```

## 8. Implementing the Parser (`parser/parser.go`)

Follow `lld_phase1_parser.md`.

```go
// parser/parser.go
package parser

import (
    "bytes"
    "fmt"
    "io"
    "net/url"
    "strings"
    "sync"

    "golang.org/x/net/html"
    "gocrawler_mvp/common"
    // "gocrawler_mvp/urlmanager" // For signaling task completion
)

// RunWorker is the main loop for a parser worker goroutine.
func RunWorker(
    id int,
    fetchedDataChan <-chan common.FetchedPageData, // From Fetcher
    newLinksChan chan<- []string,                 // To URLManager
    pageStoreChan chan<- common.PageStorageData,   // To StorageAdapter
    urlMgrTaskCompleter func(), // Function to call to signal URLManager task is done
    shutdownChan <-chan struct{},                  // From main or URLManager
    wg *sync.WaitGroup,
) {
    defer wg.Done()
    fmt.Printf("[Parser %d] Starting\n", id)

    for {
        select {
        case fetchedData, ok := <-fetchedDataChan:
            if !ok {
                fmt.Printf("[Parser %d] Fetched data channel closed. Exiting.\n", id)
                return
            }

            // Always signal task completion for the URL that was fetched,
            // regardless of parsing outcome.
            defer urlMgrTaskCompleter()

            if fetchedData.Error != nil {
                fmt.Printf("[Parser %d] Skipping parse for %s due to fetch error: %v\n", id, fetchedData.URL, fetchedData.Error)
                continue // Next item from channel
            }

            fmt.Printf("[Parser %d] Parsing: %s\n", id, fetchedData.URL)
            
            pageBaseURL, err := url.Parse(fetchedData.URL)
            if err != nil {
                fmt.Printf("[Parser %d] Error parsing base URL %s: %v\n", id, fetchedData.URL, err)
                continue
            }

            doc, err := html.Parse(bytes.NewReader(fetchedData.Body))
            if err != nil {
                fmt.Printf("[Parser %d] Error parsing HTML from %s: %v\n", id, fetchedData.URL, err)
                continue
            }

            var extractedLinks []string
            var pageTitle string
            var f func(*html.Node)

            f = func(n *html.Node) {
                if n.Type == html.ElementNode {
                    if n.Data == "title" {
                        if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
                            if pageTitle == "" { // Take first title
                                pageTitle = strings.TrimSpace(n.FirstChild.Data)
                            }
                        }
                    } else if n.Data == "a" {
                        for _, attr := range n.Attr {
                            if attr.Key == "href" {
                                linkVal := strings.TrimSpace(attr.Val)
                                if linkVal == "" || strings.HasPrefix(linkVal, "#") || strings.HasPrefix(strings.ToLower(linkVal), "javascript:") || strings.HasPrefix(strings.ToLower(linkVal), "mailto:") {
                                    continue
                                }
                                
                                parsedLink, err := url.Parse(linkVal)
                                if err == nil {
                                    absLink := pageBaseURL.ResolveReference(parsedLink)
                                    extractedLinks = append(extractedLinks, absLink.String())
                                }
                                break
                            }
                        }
                    }
                }
                for c := n.FirstChild; c != nil; c = c.NextSibling {
                    f(c)
                }
            }
            f(doc)

            // Send extracted links to URL Manager
            if len(extractedLinks) > 0 {
                select {
                case newLinksChan <- extractedLinks:
                case <-shutdownChan:
                    fmt.Printf("[Parser %d] Shutdown while sending new links. Exiting.\n", id)
                    return
                }
            }

            // Send page data to Storage Adapter
            storageData := common.PageStorageData{URL: fetchedData.URL, Title: pageTitle}
            select {
            case pageStoreChan <- storageData:
            case <-shutdownChan:
                fmt.Printf("[Parser %d] Shutdown while sending to storage. Exiting.\n", id)
                return
            }
            fmt.Printf("[Parser %d] Finished parsing %s. Title: '%s', Links: %d\n", id, fetchedData.URL, pageTitle, len(extractedLinks))

        case <-shutdownChan:
            fmt.Printf("[Parser %d] Shutdown signal received. Exiting.\n", id)
            return
        }
    }
}

```

## 9. Implementing the Storage Adapter (`storage/storage.go`)

Follow `lld_phase1_storage_postgres.md`.

```go
// storage/storage.go
package storage

import (
    "database/sql"
    "fmt"
    "sync"
    "time"

    _ "github.com/lib/pq" // PostgreSQL driver
    "gocrawler_mvp/common"
    "gocrawler_mvp/config"
)

type Adapter struct {
    db *sql.DB
}

func NewAdapter(cfg *config.Manager) (*Adapter, error) {
    connStr := cfg.GetDBConnectionString()
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, fmt.Errorf("failed to open DB connection: %w", err)
    }

    err = db.Ping()
    if err != nil {
        db.Close() // Close if ping fails
        return nil, fmt.Errorf("failed to ping DB: %w", err)
    }

    // Optional: Set connection pool parameters
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(time.Hour)

    fmt.Println("[Storage] Successfully connected to PostgreSQL.")
    return &Adapter{db: db}, nil
}

func (a *Adapter) SavePage(data common.PageStorageData) error {
    sqlStatement := `INSERT INTO pages (url, title) VALUES ($1, $2) ON CONFLICT (url) DO NOTHING;`
    _, err := a.db.Exec(sqlStatement, data.URL, data.Title)
    if err != nil {
        return fmt.Errorf("failed to save page %s: %w", data.URL, err)
    }
    // fmt.Printf("[Storage] Saved page: %s\n", data.URL)
    return nil
}

func (a *Adapter) Close() error {
    if a.db != nil {
        fmt.Println("[Storage] Closing database connection.")
        return a.db.Close()
    }
    return nil
}

// RunStorageWorker is the main loop for a storage worker goroutine.
func RunStorageWorker(
    adapter *Adapter,
    pageStoreChan <-chan common.PageStorageData, // From Parser
    urlMgrTaskCompleter func(), // To signal URLManager that this part of the task is done for its counting
                                // Note: The original LLD for parser implied parser calls TaskCompleted.
                                // If storage is the final step for a URL's data, then storage worker might call it.
                                // For now, let's assume parser calls TaskCompleted after it has successfully sent data to pageStoreChan.
                                // The urlMgrTaskCompleter is not used here under that assumption.
    shutdownChan <-chan struct{},                  // From main
    wg *sync.WaitGroup,
) {
    defer wg.Done()
    fmt.Println("[StorageWorker] Starting")

    for {
        select {
        case pageData, ok := <-pageStoreChan:
            if !ok {
                fmt.Println("[StorageWorker] Page store channel closed. Exiting.")
                return
            }
            err := adapter.SavePage(pageData)
            if err != nil {
                fmt.Printf("[StorageWorker] Error saving page %s: %v\n", pageData.URL, err)
            } else {
                fmt.Printf("[StorageWorker] Successfully processed for storage: %s\n", pageData.URL)
            }
            // If storage was the one to complete the task for URLManager:
            // urlMgrTaskCompleter() 

        case <-shutdownChan:
            fmt.Println("[StorageWorker] Shutdown signal received. Exiting.")
            return
        }
    }
}

```

## 10. Integrating Components (`cmd/main.go`)

This is where everything comes together.

```go
// cmd/main.go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "sync"
    "syscall"

    "gocrawler_mvp/cli"
    "gocrawler_mvp/common"
    "gocrawler_mvp/config"
    "gocrawler_mvp/fetcher"
    "gocrawler_mvp/parser"
    "gocrawler_mvp/storage"
    "gocrawler_mvp/urlmanager"
)

func main() {
    // 1. Parse CLI Flags
    cliFlags, err := cli.ParseFlags()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
        os.Exit(1)
    }

    // 2. Initialize Configuration Manager
    cfg := config.NewManager(
        cliFlags.SeedURLsParsed,
        cliFlags.NumWorkers,
        cliFlags.CrawlDelayMs,
        cliFlags.DBConnectionString, // Use parsed or default from config.NewManager
        cliFlags.UserAgent,      // Use parsed or default from config.NewManager
    )
    fmt.Printf("[Main] Configuration loaded. Workers: %d, Delay: %v\n", cfg.GetNumWorkers(), cfg.GetCrawlDelay())

    // 3. Initialize Storage Adapter
    store, err := storage.NewAdapter(cfg)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error initializing storage adapter: %v\n", err)
        os.Exit(1)
    }
    defer store.Close()

    // 4. Initialize URL Manager
    urlMgr := urlmanager.NewManager()
    go urlMgr.CheckDone() // Goroutine to close urlChan when all tasks are done

    // 5. Setup Channels
    // urlMgr.urlChan is used by fetchers
    fetchedDataChan := make(chan common.FetchedPageData, cfg.GetNumWorkers()*2) // Buffer slightly larger
    newLinksChan := make(chan []string, cfg.GetNumWorkers()*2)
    pageStoreChan := make(chan common.PageStorageData, cfg.GetNumWorkers()*2)

    // 6. Start Workers (Fetchers, Parsers, StorageWorker)
    var workerWg sync.WaitGroup

    // Start Fetcher Workers
    for i := 0; i < cfg.GetNumWorkers(); i++ {
        workerWg.Add(1)
        go fetcher.RunWorker(i+1, cfg, urlMgr.GetNextURL, fetchedDataChan, urlMgr.SignalShutdown, &workerWg)
    }

    // Start Parser Workers
    for i := 0; i < cfg.GetNumWorkers(); i++ {
        workerWg.Add(1)
        go parser.RunWorker(i+1, fetchedDataChan, newLinksChan, pageStoreChan, urlMgr.TaskCompleted, urlMgr.SignalShutdown, &workerWg)
    }
    
    // Start a single Storage Worker (or a pool if DB becomes a bottleneck)
    workerWg.Add(1)
    go storage.RunStorageWorker(store, pageStoreChan, urlMgr.SignalShutdown, &workerWg) // Removed urlMgrTaskCompleter for now

    // 7. Add Seed URLs
    urlMgr.AddURLs(cfg.GetSeedURLs(), nil) // nil for baseURL as these are absolute seeds

    // 8. Handle Graceful Shutdown (Ctrl+C)
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // Goroutine to listen for new links and add them to URL Manager
    // And also to listen for shutdown signals to stop adding links.
    go func() {
        for {
            select {
            case links := <-newLinksChan:
                // For links from parser, the base URL is implicit in how they were resolved.
                // However, AddURLs expects a baseURL for resolving relative ones if any slip through.
                // This part needs careful thought: parser should send absolute URLs.
                // If parser sends absolute URLs, baseURL for AddURLs can be nil.
                urlMgr.AddURLs(links, nil) 
            case <-urlMgr.SignalShutdown: // Listen to the URL Manager's shutdown signal
                fmt.Println("[Main] URL Manager shutdown detected, stopping link listener.")
                return
            }
        }
    }()

    fmt.Println("[Main] Crawler started. Press Ctrl+C to stop.")

    // Wait for either OS signal or for the URL manager to be done
    select {
    case s := <-sigChan:
        fmt.Printf("\n[Main] Received signal: %s. Shutting down...\n", s)
        urlMgr.SignalShutdown()
    case <-urlMgr.SignalShutdown: // If CheckDone closes it first
        fmt.Println("\n[Main] Crawler finished its work. Shutting down.")
        // urlMgr.SignalShutdown() // already called by CheckDone implicitly or explicitly
    }
    
    // Wait for all worker goroutines to finish
    fmt.Println("[Main] Waiting for workers to complete...")
    workerWg.Wait()
    fmt.Println("[Main] All workers finished. Exiting.")
}

```

## 11. Building and Running

1.  **Get Dependencies:**
    ```bash
    go get github.com/lib/pq 
    go mod tidy 
    ```
2.  **Build:**
    ```bash
    go build -o gocrawler_mvp ./cmd/main.go
    ```
3.  **Run (ensure PostgreSQL is running and configured):**
    ```bash
    ./gocrawler_mvp -seeds "http://books.toscrape.com" -workers 3 -delay 1000 -dbconn "postgres://youruser:yourpass@localhost:5432/crawler_mvp_db?sslmode=disable"
    ```
    Replace the connection string and other flags as needed.

## 12. Next Steps & Refinements

*   **Thorough Testing:** Test each component and the integrated system.
*   **Logging:** Implement more structured logging (e.g., using a logging library).
*   **Error Handling:** Enhance error handling and reporting.
*   **Metrics:** Add more detailed metrics (pages/sec, error rates).
*   **Refine `urlMgr.TaskCompleted` logic:** Ensure it's called correctly and only once per URL lifecycle to accurately track `activeTasks`.
    *   A URL is truly "done" for the `URLManager`'s `activeTasks` count when it has been fetched, parsed, its data stored, and its new links (if any) have been added back to the `URLManager`. The current `parser.RunWorker` calls `urlMgr.TaskCompleted()`. This means the task is marked complete before its extracted links are processed by the `urlMgr.AddURLs` via `newLinksChan`. This could lead to a premature shutdown if `activeTasks` hits zero while links are still in `newLinksChan`. This needs refinement. A better approach might be for `URLManager` to decrement `activeTasks` only after it has processed the links from a given URL *and* received confirmation that its data has been sent to storage.

This guide provides a comprehensive starting point. You will likely encounter areas for refinement as you implement and test. Good luck!
