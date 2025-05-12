# Distributed Web Crawler in Go: From Design to Implementation

This document outlines the steps to build a distributed web crawler using Go. It covers project setup, core functionalities, and best practices for creating a robust and scalable crawler. **Crucially, this guide emphasizes starting with High-Level Design (HLD) and Low-Level Design (LLD) before diving into code, and refers to the separate HLD and LLD documents you will create.**

## 0. Design First: HLD and LLD

Before writing any Go code, it is highly recommended to complete your High-Level Design (HLD) and Low-Level Design (LLD) documents. These will serve as your blueprints.

*   **Refer to `hld_template.md`:** Use this template to create your overall system architecture for the web crawler. Think about the main components, their interactions, data flow, scalability, and fault tolerance.
*   **Refer to `lld_fetcher_template.md`, `lld_parser_template.md`, and `lld_scheduler_template.md`:** Use these templates to detail the design of each core component. This includes defining responsibilities, interfaces, data structures, core logic, and error handling for the Fetcher, Parser, and Scheduler/URL Frontier.

**The implementation steps below should be guided by the decisions you make in your HLD and LLD documents.**

## 1. Project Setup

*   **Initialize Go module:** `go mod init crawler` (as per your HLD/LLD considerations for project naming).
*   **Create project structure (refer to your LLD for package structure):**
    ```
    crawler/  <-- Your project root, named as per your HLD
    ├── main.go
    ├── hld_crawler.md  <-- Your completed HLD document
    ├── lld/
    │   ├── lld_fetcher.md
    │   ├── lld_parser.md
    │   └── lld_scheduler.md
    ├── fetcher/          <-- Package for Fetcher component (as per LLD)
    │   └── fetcher.go
    ├── parser/           <-- Package for Parser component (as per LLD)
    │   └── parser.go
    ├── scheduler/        <-- Package for Scheduler/URL Frontier (as per LLD)
    │   └── scheduler.go
    ├── storage/          <-- Package for Storage (as per LLD, if separated)
    │   └── storage.go
    ├── common/           <-- For shared data structures (e.g., CrawlTask, FetchedPage)
    │   └── types.go
    ├── utils/            <-- Utility functions
    │   └── utils.go
    └── README.md         <-- This implementation guide, or your main project README
    ```

## 2. Core Functionalities (Derived from HLD)

Your HLD document will define these. Key functionalities typically include:

*   **URL Frontier Management:** As detailed in your Scheduler LLD.
*   **Page Fetching:** As detailed in your Fetcher LLD.
*   **Content Parsing:** As detailed in your Parser LLD.
*   **Data Storage:** Design based on HLD and potentially a separate LLD if complex.
*   **Scheduling and Coordination:** As detailed in your Scheduler LLD.
*   **Concurrency:** A core aspect of your HLD and LLD for each component.

## 3. Implementation Steps (Guided by your LLDs)

### Step 1: Define Core Data Structures (refer to `common/types.go` and LLDs)

*   Based on your LLDs, define shared structures like `CrawlTask`, `FetchedPage`, `ParsedData`, etc., in a `common` package.

### Step 2: Implement Core Components (One by one, following your LLDs)

*   **Fetcher (`fetcher/fetcher.go`):**
    *   Implement the Fetcher component as designed in `lld_fetcher.md`.
    *   Focus on HTTP requests, response handling, retries, and respecting `robots.txt` as per your LLD.
*   **Parser (`parser/parser.go`):**
    *   Implement the Parser component as designed in `lld_parser.md`.
    *   Focus on HTML parsing, link extraction, URL normalization, and data extraction rules from your LLD.
*   **Scheduler/URL Frontier (`scheduler/scheduler.go`):**
    *   Implement the Scheduler/URL Frontier as designed in `lld_scheduler.md`.
    *   Focus on queue management, politeness, duplicate detection, and providing URLs to Fetchers, all based on your LLD.
*   **Storage (`storage/storage.go`):**
    *   Implement the storage solution chosen in your HLD and detailed in its LLD (if created).

### Step 3: Implement Concurrency (Refer to HLD and LLDs)

*   **Worker Pools:** Based on your HLD/LLD, create worker pools for Fetchers and Parsers.
*   **Channel Communication:** Use channels for communication between components as designed in your LLDs (e.g., Scheduler to Fetchers, Fetchers to Parsers, Parsers back to Scheduler).

### Step 4: Build the Main Application (`main.go`)

*   Initialize all components based on their LLDs and configurations.
*   Start the worker pools and the main crawling loop as per your HLD.
*   Implement any monitoring or control mechanisms defined in your HLD.

## 4. Advanced Features (Optional - Consider for HLD v2 / LLD v2)

*   **Robots.txt Respect:** Ensure this is deeply integrated as per your Fetcher/Scheduler LLD.
*   **Duplicate Content Detection:** As designed in your HLD/LLD.
*   **Distributed Coordination:** If your HLD aims for a truly distributed system, this is a major implementation phase.

## 5. Deployment and Execution

*   **Build the application:** `go build`
*   **Run the application:** `./crawler` (or your chosen executable name)

This guide now strongly encourages you to refer to your HLD and LLD documents at each stage. This practice of designing before coding is key to building complex systems.

## 6. GitHub and Resume Presentation

*   **Clear README:** Your main project `README.md` is crucial. It should include:
    *   Project Title and Description.
    *   **Link to your HLD and LLD documents.**
    *   Features Implemented.
    *   Technologies Used (Go, specific libraries).
    *   Setup and Installation Instructions.
    *   Usage Examples.
    *   Project Structure Explanation.
*   **Well-structured Code:** Follow Go conventions and the structure defined in your LLDs.
*   **Comments:** Add meaningful comments, especially explaining how the code implements LLD decisions.
*   **Unit Tests:** Write unit tests for key functions and methods defined in your LLDs.
*   **Makefile (Optional but good):** A Makefile can simplify build, test, and run commands.
*   **Git Commits:** Use clear and descriptive commit messages, possibly referencing HLD/LLD items.
*   **Demonstrate Go Internals Understanding (in README or code comments):**
    *   Explain how you used goroutines and channels for concurrency, as per your LLDs.
*   **Resume Points:**
    *   Highlight that you designed the system using HLD and LLD principles before implementation.
    *   Clearly state the project and its purpose.
    *   Mention challenges faced during design and implementation and how you overcame them.
    *   Link to your GitHub repository where HLD/LLD docs are visible.

By following this design-first approach, you will not only build a more robust web crawler but also gain valuable experience in software design, which is highly sought after.
