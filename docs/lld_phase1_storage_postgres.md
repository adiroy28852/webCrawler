# Low-Level Design (LLD): Phase 1 - Storage Adapter (PostgreSQL)

## 1. Introduction

*   **Purpose:** This document details the Low-Level Design for the Storage Adapter component of the Phase 1 MVP web crawler, specifically for interacting with a PostgreSQL database.
*   **Scope:** Defines how crawled page information (URL and title for MVP) is persisted to the PostgreSQL database.
*   **Reference HLD:** `hld_phase1_mvp.md`
*   **Reference LLD (Parser):** `lld_phase1_parser.md` (provides input to Storage Adapter)
*   **Reference LLD (Config Manager):** `lld_phase1_config_manager.md` (provides DB connection string)

## 2. Responsibilities

*   Establish and manage a connection to the PostgreSQL database using the connection string from the Configuration Manager.
*   Receive page data (URL, title) from Parser workers (via a channel).
*   Save this page data into the `pages` table in the PostgreSQL database.
*   Handle potential database errors gracefully (e.g., connection issues, SQL errors like unique constraint violations).

## 3. Interfaces

*   **Input:**
    *   Receives `PageStorageData` from a channel connected to Parser workers.
    ```go
    // From common package or parser LLD
    type PageStorageData struct {
        URL   string
        Title string
    }
    ```
    *   Accesses the database connection string from the Configuration Manager.
*   **Output:**
    *   No direct data output to other components, but logs success or failure of database operations.
*   **Functions/Methods (Conceptual):**
    *   `NewAdapter(config *config.Manager) (*StorageAdapter, error)`: Constructor to initialize the adapter and establish DB connection.
    *   `SavePage(data PageStorageData) error`: Method to save a single page's data.
    *   `RunStorageWorker(pageStoreChan <-chan PageStorageData, config *config.Manager, shutdownChan <-chan struct{}, wg *sync.WaitGroup)`: The main loop for a dedicated storage worker goroutine (recommended to decouple DB operations).
    *   `Close() error`: Method to close the database connection gracefully.

## 4. Data Structures

*   ```go
    // Package: storage
    import (
        "database/sql"
        _ "github.com/lib/pq" // PostgreSQL driver
        "sync"
        // other imports like config, common
    )

    type Adapter struct {
        db *sql.DB
        // config *config.Manager // If needed directly for more than just conn string
    }
    ```
*   The `PageStorageData` struct (defined in Parser LLD or a common package) is used for input.

## 5. Core Logic / Algorithms

*   **Initialization (`NewAdapter` function):**
    1.  Get DB connection string from `config.GetDBConnectionString()`.
    2.  Open a database connection: `db, err := sql.Open("postgres", connString)`.
    3.  If `err != nil`, return `nil, err`.
    4.  Ping the database to verify the connection: `err = db.Ping()`.
    5.  If `err != nil`, close `db` and return `nil, err`.
    6.  Set connection pool parameters (optional for MVP, but good practice):
        *   `db.SetMaxOpenConns(10)` (example)
        *   `db.SetMaxIdleConns(5)` (example)
        *   `db.SetConnMaxLifetime(time.Hour)` (example)
    7.  Return `&Adapter{db: db}, nil`.

*   **`RunStorageWorker` (Dedicated Goroutine for DB Writes):**
    1.  Worker loop: `for { select { case pageData, ok := <-pageStoreChan: ... case <-shutdownChan: return } }`
    2.  If `pageStoreChan` is closed (`!ok`), the worker terminates.
    3.  Call `adapter.SavePage(pageData)`.
    4.  Log success or error from `SavePage`.

*   **`SavePage(data PageStorageData) error`:**
    1.  SQL Statement: `INSERT INTO pages (url, title) VALUES ($1, $2) ON CONFLICT (url) DO NOTHING;`
        *   `ON CONFLICT (url) DO NOTHING` handles cases where a URL might be processed by a parser and sent for storage multiple times due to concurrent processing or retries. This ensures we don't get unique constraint errors if the URL is already there.
    2.  Execute the statement: `_, err := a.db.Exec(sqlStatement, data.URL, data.Title)`.
    3.  If `err != nil`, log the error (e.g., "Failed to save page [URL]: [Error]") and return `err`.
    4.  If successful, log (e.g., "Successfully saved page [URL]") and return `nil`.

*   **Database Schema (reiteration from HLD):**
    ```sql
    CREATE TABLE IF NOT EXISTS pages (
        id SERIAL PRIMARY KEY,
        url TEXT NOT NULL UNIQUE,
        title TEXT,
        crawled_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    ```
    *   The application should ensure this table exists. For MVP, this might be a manual setup step for the user, or the application could attempt to create it if it doesn't exist (more complex error handling).

*   **`Close() error`:**
    1.  `return a.db.Close()`

## 6. Error Handling

*   **DB Connection Errors (`NewAdapter`):** Logged and returned, preventing the adapter from being created. The main application should handle this (e.g., exit).
*   **SQL Execution Errors (`SavePage`):**
    *   Logged.
    *   For `ON CONFLICT...DO NOTHING`, a unique constraint violation is not an error but an expected outcome if the URL was already saved. `db.Exec` would not return an error in this specific case for PostgreSQL.
    *   Other SQL errors (e.g., table not found, syntax errors in query, connection lost during exec) should be logged. The application might choose to continue or implement a circuit breaker for DB issues.
*   **Channel Send/Receive Errors:** Not applicable if using standard channel operations correctly.

## 7. Concurrency Considerations

*   The `database/sql` package in Go provides a connection pool that is safe for concurrent use. Multiple goroutines can call methods like `Exec` or `Query` on the same `*sql.DB` instance.
*   Running database write operations in one or a small pool of dedicated `StorageWorker` goroutines can help manage database load and centralize DB interaction logic, rather than having many Parser workers directly interact with the DB.
*   The `SavePage` method itself is stateless and operates on the input data.

## 8. Dependencies

*   Go standard library: `database/sql`, `fmt`, `time` (for connection pool settings).
*   PostgreSQL driver: `github.com/lib/pq` (imported with `_` for side effects).
*   Internal: Configuration Manager (for DB connection string), Parser (for input `PageStorageData`).

## 9. Design Choices & Reasoning

*   **PostgreSQL:** A robust, open-source relational database, suitable for structured data like crawled page information. User expressed interest.
*   **`database/sql` package:** Go's standard interface for SQL databases, promoting portable code (though the driver is specific).
*   **`lib/pq` driver:** A widely used and stable PostgreSQL driver for Go.
*   **`ON CONFLICT (url) DO NOTHING`:** A simple and effective way to handle potential duplicate save attempts for the MVP, leveraging PostgreSQL's features. This avoids needing to query `IF NOT EXISTS` before inserting, which is less atomic.
*   **Dedicated Storage Worker Goroutine(s):** Decouples parsing logic from direct database writes. This can be beneficial for managing backpressure if the database is slow, and centralizes DB write logic.
*   **Connection Pooling:** Handled by `database/sql`, which is efficient and standard.

This LLD for the Storage Adapter provides a clear plan for persisting crawled data to PostgreSQL in the MVP.
