# High-Level Design (HLD): Phase 1 - MVP CLI-based Web Crawler

## 1. Introduction

*   **Purpose:** This document outlines the High-Level Design for a Minimum Viable Product (MVP) of a web crawler. The MVP will be a single-node application, operated via a Command Line Interface (CLI), and will use PostgreSQL for data storage.
*   **Scope:** The crawler will accept seed URLs, fetch web pages, extract new links from these pages, and store basic information about the crawled pages (URL, title, crawl timestamp).
*   **Goals for Phase 1 MVP:**
    *   Create a functional, runnable web crawler.
    *   Provide a learning platform for Go fundamentals, concurrency (goroutines and channels), basic web interactions (HTTP requests, HTML parsing), and PostgreSQL integration.
    *   Establish a foundation for future enhancements in Phase 2 (Cloud Deployment) and Phase 3 (Large Scale).
*   **Non-Goals for Phase 1 MVP:**
    *   Distributed crawling across multiple machines.
    *   Advanced politeness mechanisms (e.g., per-domain adaptive crawl delays, `robots.txt` parsing beyond a simple check if we decide to add it).
    *   Complex User Interface (UI) beyond CLI interactions.
    *   Large-scale performance optimizations or handling massive datasets.
    *   JavaScript rendering.

## 2. System Architecture Overview

*   **Architecture Style:** Single-node, concurrent application.
*   **User Interaction:** Via Command Line Interface (CLI) for initiating crawls, providing seed URLs, and viewing basic status updates.
*   **Core Internal Components:** The system will be composed of several distinct components working together:
    1.  **CLI Interface:** Handles user input and displays output.
    2.  **Configuration Manager:** Manages application settings.
    3.  **URL Manager (Frontier):** Manages the queue of URLs to be crawled and tracks visited URLs.
    4.  **Fetcher Workers:** A pool of workers responsible for downloading web page content.
    5.  **Parser Workers:** A pool of workers responsible for parsing HTML content and extracting links/data.
    6.  **Storage Adapter (PostgreSQL):** Handles all interactions with the PostgreSQL database.
*   **Concurrency Model:** The system will leverage Go's goroutines for concurrent execution of Fetcher and Parser tasks. Channels will be used for communication and data flow between components.

## 3. Component Descriptions

*   **CLI Interface:**
    *   **Responsibilities:** Parse command-line arguments (e.g., seed URLs, number of workers), initiate the crawling process, display progress information (e.g., number of pages crawled, queue size), and handle termination signals.
    *   **Interaction:** Passes initial seed URLs to the URL Manager. Triggers the main crawl controller/scheduler logic.
*   **Configuration Manager:**
    *   **Responsibilities:** Load and provide access to application settings. For the MVP, these will be a combination of hardcoded defaults and optional overrides via CLI flags.
    *   **Settings Example:** Seed URLs, number of fetcher workers, number of parser workers, default crawl delay (a simple global delay for MVP), PostgreSQL connection string, User-Agent string.
    *   **Interaction:** Other components will request configuration values from this manager.
*   **URL Manager (Frontier):**
    *   **Responsibilities:** Maintain a queue of URLs to be crawled. Provide the next available URL to Fetcher workers. Prevent duplicate crawling of the same URL within a single crawl session. For MVP, this will be an in-memory solution.
    *   **Data Structures (MVP):**
        *   A buffered channel for the URL queue to allow concurrent access.
        *   A map (`map[string]bool`) to store visited/processed URLs for duplicate checking.
    *   **Interaction:** Receives seed URLs from the CLI. Receives newly discovered URLs from Parser workers. Provides URLs to Fetcher workers.
*   **Fetcher Workers (Pool):**
    *   **Responsibilities:** Each worker takes a URL from the URL Manager. Makes an HTTP GET request to fetch the page content. Handles basic HTTP errors (e.g., timeouts, non-200 status codes). Implements a simple retry mechanism for transient network errors.
    *   **Technology:** Uses Go's standard `net/http` package.
    *   **Configuration:** Uses a configurable User-Agent string.
    *   **Interaction:** Retrieves URLs from the URL Manager. Passes the fetched HTML content (and the URL it was fetched from) to a Parser worker.
*   **Parser Workers (Pool):**
    *   **Responsibilities:** Each worker takes raw HTML content (and its source URL) from a Fetcher worker. Parses the HTML to extract new hyperlinks. For MVP, it will also extract the page title.
    *   **URL Normalization:** Converts relative URLs to absolute URLs. Performs basic normalization (e.g., ensuring scheme is present).
    *   **Technology:** Uses Go's standard `golang.org/x/net/html` package for parsing.
    *   **Interaction:** Receives HTML content from Fetcher workers. Sends newly discovered and normalized URLs to the URL Manager. Sends extracted page data (URL, title) to the Storage Adapter.
*   **Storage Adapter (PostgreSQL):**
    *   **Responsibilities:** Encapsulate all database interactions. Connect to the PostgreSQL database. Save information about crawled pages.
    *   **Technology:** Uses Go's `database/sql` package with the `lib/pq` driver for PostgreSQL.
    *   **Schema (MVP):** A simple `pages` table:
        ```sql
        CREATE TABLE IF NOT EXISTS pages (
            id SERIAL PRIMARY KEY,
            url TEXT NOT NULL UNIQUE,
            title TEXT,
            crawled_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );
        ```
    *   **Interaction:** Receives page data (URL, title) from Parser workers to be saved.

## 4. Data Flow

1.  **Initialization:** User starts the application via CLI, providing seed URLs and optional configuration overrides (e.g., number of workers).
2.  **Configuration Loaded:** Configuration Manager loads settings.
3.  **Seed URLs:** CLI passes seed URLs to the URL Manager, which adds them to its queue if not already visited.
4.  **Worker Dispatch:** The main application logic (acting as a simple scheduler) starts a configurable number of Fetcher and Parser goroutines (workers).
5.  **URL Request:** A Fetcher worker requests a URL from the URL Manager.
6.  **URL Provision:** If the URL queue is not empty, the URL Manager provides a URL to the Fetcher worker and marks it as "in-progress" (or removes it from the immediate queue and adds to visited set to prevent re-queuing by other parsers before it's fully processed).
7.  **Page Fetching:** The Fetcher worker makes an HTTP GET request. Upon success, it passes the HTML content and the source URL to a Parser worker (e.g., via a channel).
8.  **Content Parsing:** The Parser worker receives the HTML. It extracts all hyperlinks and the page title.
9.  **Link Processing:** Extracted links are normalized. The Parser worker sends these new, normalized URLs to the URL Manager. The URL Manager checks for duplicates (against its visited set) and adds new, unique URLs to its queue.
10. **Data Storage:** The Parser worker sends the crawled page's URL and extracted title to the Storage Adapter, which saves it to the PostgreSQL `pages` table.
11. **Loop:** Fetcher and Parser workers continue processing until the URL Manager's queue is empty and all active tasks are completed, or a user-defined limit (e.g., max pages) is reached.
12. **Termination:** The application reports final statistics and exits.

## 5. Key Design Considerations (MVP Focus)

*   **Scalability:** The MVP is designed for a single node. Scalability is a non-goal for Phase 1 but will be addressed in later phases.
*   **Fault Tolerance:** Basic error handling will be implemented within Fetcher workers (e.g., retrying failed HTTP requests a couple of times). If a worker goroutine panics, it will be logged, and the specific task might fail, but the overall application should aim to continue if other workers are available. For MVP, a panic in a worker might just terminate that worker.
*   **Concurrency Model:**
    *   A fixed pool of Fetcher goroutines.
    *   A fixed pool of Parser goroutines.
    *   Channels for inter-component communication:
        *   `urls_to_fetch_chan`: From URL Manager to Fetchers.
        *   `html_to_parse_chan`: From Fetchers to Parsers.
        *   `new_links_chan`: From Parsers to URL Manager.
        *   `data_to_store_chan`: From Parsers to Storage Adapter.
    *   Use of WaitGroups to manage the lifecycle of worker pools and ensure graceful shutdown.
*   **Politeness:** A simple, global, configurable delay will be implemented between consecutive HTTP requests made by *any* Fetcher worker. This is a very basic mechanism for MVP. Per-domain politeness is a Phase 2+ feature.
*   **Data Storage Strategy:** PostgreSQL, as detailed in the Storage Adapter component. Connection pooling will be handled by the `database/sql` package.
*   **Duplicate URL Handling:** An in-memory `map[string]bool` in the URL Manager will track all URLs that have been added to the queue or successfully crawled to prevent redundant processing within the same crawl session.
*   **Configuration:** Primarily through CLI flags for key parameters (seed URLs, worker counts, delay), with sensible hardcoded defaults.

## 6. Technology Choices (Prescribed for MVP)

*   **Programming Language:** Go
*   **HTTP Client:** Go standard library (`net/http`)
*   **HTML Parsing:** Go standard library (`golang.org/x/net/html`)
*   **Database Interaction:** Go standard library (`database/sql`) with the `lib/pq` PostgreSQL driver.
*   **Command-Line Arguments:** Go standard library (`flag`)

## 7. Assumptions and Dependencies

*   A PostgreSQL server must be running and accessible to the application with appropriate credentials.
*   The machine running the crawler must have internet connectivity.
*   Seed URLs are valid and accessible.

## 8. Transition to Phase 2 (Cloud) & Phase 3 (Large Scale) - Brief Outlook

This MVP design lays the groundwork for future scaling:

*   **Phase 2 (Cloud Deployment):**
    *   **URL Manager:** The in-memory queue can be replaced with a cloud-based message queue like AWS SQS for better decoupling and persistence.
    *   **Workers:** Fetcher and Parser logic can be packaged into Docker containers and deployed on services like AWS ECS, EKS, or potentially as Lambda functions (though long-running crawlers might be better suited for container services).
    *   **Storage:** Transition to a managed database service like AWS RDS for PostgreSQL.
    *   **Configuration:** Utilize cloud-based configuration management (e.g., AWS Systems Manager Parameter Store).
    *   **Politeness:** Implement more robust per-domain politeness, possibly using a distributed cache like Redis to store last access times for domains.
*   **Phase 3 (Large Scale):**
    *   **Distributed URL Frontier:** Explore more advanced distributed queueing systems (e.g., Kafka) or custom solutions for managing a massive number of URLs and ensuring efficient distribution.
    *   **Duplicate Detection:** Implement Bloom filters or similar probabilistic data structures, potentially backed by a persistent key-value store, for highly scalable duplicate detection.
    *   **Coordination & Scheduling:** Introduce dedicated coordinator services for managing worker nodes, distributing tasks intelligently, and handling failures in a distributed environment.
    *   **Monitoring & Logging:** Integrate comprehensive monitoring and logging solutions (e.g., Prometheus, Grafana, ELK stack).

This HLD for Phase 1 provides a clear, prescriptive starting point. The next step will be to create detailed Low-Level Designs (LLDs) for each component based on this HLD.
