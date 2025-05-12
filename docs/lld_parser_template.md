# Low-Level Design (LLD): Parser Component

## 1. Introduction

*   **Purpose:** Describe the responsibility of the Parser component in the web crawler.
*   **Scope:** What this component does and does not do.

## 2. Responsibilities

*   Receives fetched page content (e.g., HTML) from the Fetcher component.
*   Parses the HTML content to extract:
    *   New URLs (links) to be crawled.
    *   Relevant data from the page (e.g., text content, specific tags, metadata), based on project requirements.
*   Normalizes and validates extracted URLs (e.g., resolves relative URLs to absolute, checks for valid schemes).
*   Filters out unwanted URLs (e.g., based on domain, file type, depth).
*   Returns the extracted links and data.

## 3. Interfaces

*   **Input:**
    *   `FetchedPage` (or similar struct): Contains the raw page content (e.g., `[]byte` or `string`), the base URL of the page (for resolving relative links), and potentially other metadata.
*   **Output:**
    *   `ParsedData` (or similar struct): Contains a list of newly discovered (and normalized) URLs, and the extracted data from the page.
    *   Any errors encountered during parsing.
*   **Functions/Methods (Conceptual):**
    *   `Parse(page FetchedPage) (ParsedData, error)`: Main function to perform the parsing operation.

## 4. Data Structures

*   `ParserConfig`: Struct to hold configuration like allowed/disallowed URL patterns, data extraction rules (e.g., CSS selectors).
*   Internal data structures for managing extracted links and data during processing.

## 5. Core Logic / Algorithms

*   **Parsing Flow:**
    1.  Receive `FetchedPage`.
    2.  Select an HTML parsing library (e.g., Go's `html` package, or a third-party library like `goquery`).
    3.  Parse the HTML document tree.
    4.  Iterate through relevant HTML elements (e.g., `<a>` tags for links, other tags for data based on `ParserConfig`).
    5.  For each link found:
        *   Resolve relative URLs to absolute URLs using the page's base URL.
        *   Normalize the URL (e.g., lowercase scheme and host, remove fragments if not needed).
        *   Validate and filter the URL based on configured rules.
        *   Add valid, new URLs to a list.
    6.  For data extraction:
        *   Apply configured selectors/rules to extract desired information.
        *   Store extracted data.
    7.  Prepare and return `ParsedData`.
*   **URL Normalization and Filtering Logic:** Detail the specific steps and rules.
*   **Data Extraction Logic:** Detail how data is identified and extracted.

## 6. Error Handling

*   List potential errors (HTML parsing errors, invalid URL formats, issues with data extraction rules).
*   How each error type is handled and reported (e.g., skip bad links, log errors, return partial data if possible).

## 7. Concurrency Considerations

*   Parsing can be CPU-intensive. If the Parser is part of a worker pool, it will run concurrently with other Parsers and Fetchers.
*   Ensure any shared data structures (e.g., if using global filter lists, though generally not recommended) are accessed in a thread-safe manner.

## 8. Dependencies

*   Go standard library: `net/url`, `strings`, `html` (or chosen parsing library).
*   External libraries: HTML parsing library (e.g., `goquery`) if not using the standard library directly.
*   Other components: Fetcher (for input), URL Frontier/Scheduler (for outputting new links), Data Storage (for outputting extracted data).

*(This is a template. You will fill in the details based on your project decisions.)*
