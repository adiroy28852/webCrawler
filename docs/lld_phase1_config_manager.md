# Low-Level Design (LLD): Phase 1 - Configuration Manager

## 1. Introduction

*   **Purpose:** This document details the Low-Level Design for the Configuration Manager component of the Phase 1 MVP web crawler.
*   **Scope:** Defines how the application loads, stores, and provides access to configuration settings. For the MVP, this will be a simple approach combining defaults with CLI overrides.
*   **Reference HLD:** `hld_phase1_mvp.md`
*   **Reference LLD (CLI):** `lld_phase1_cli.md` (as CLI flags are the source of overrides)

## 2. Responsibilities

*   Define default values for all configurable parameters.
*   Store the final configuration values (defaults overridden by CLI flags where provided).
*   Provide a centralized and consistent way for other components to access configuration settings.
*   For MVP, configuration is loaded once at startup.

## 3. Interfaces

*   **Input:**
    *   Receives parsed CLI flags (e.g., `CLIFlags` struct from `lld_phase1_cli.md`) or individual flag values during initialization.
*   **Output (Getter Methods):**
    *   `GetSeedURLs() []string`
    *   `GetNumWorkers() int`
    *   `GetCrawlDelay() time.Duration`
    *   `GetDBConnectionString() string`
    *   `GetUserAgent() string`
*   **Internal Interaction:**
    *   Initialized by the main application logic after CLI flags are parsed.
    *   Other components (URL Manager, Fetcher, Storage Adapter, etc.) will call its getter methods.

## 4. Data Structures

*   A struct to hold all configuration settings:
    ```go
    // Package: config
    type Manager struct {
        SeedURLs         []string
        NumWorkers       int
        CrawlDelay       time.Duration
        DBConnectionString string
        UserAgent        string
        // Potentially other settings for MVP if identified
    }
    ```
    *   This struct will be a singleton or a globally accessible instance (e.g., passed around or accessed via a package-level variable after initialization).

## 5. Core Logic / Algorithms

*   **Initialization (`NewManager` function or similar):**
    1.  Accepts the parsed `CLIFlags` struct (or individual values) as input.
    2.  Define hardcoded default values for each setting.
        *   Default `NumWorkers`: e.g., `3`
        *   Default `CrawlDelay`: e.g., `1 * time.Second`
        *   Default `DBConnectionString`: e.g., `"postgres://postgres:password@localhost:5432/crawler_mvp_db?sslmode=disable"` (Emphasize this is a placeholder and should be configured securely).
        *   Default `UserAgent`: e.g., `"GoCrawlerMVP/0.1 (+http://yourproject.url.here)"`
    3.  For each setting, use the value from `CLIFlags` if provided and valid; otherwise, use the default.
        *   Example for `NumWorkers`:
            ```go
            if cliFlags.NumWorkers > 0 {
                cfg.NumWorkers = cliFlags.NumWorkers
            } else {
                cfg.NumWorkers = DefaultNumWorkers // default value
            }
            ```
        *   Seed URLs are directly taken from CLI as they are required.
    4.  Store the final values in the `Manager` struct instance.
    5.  Return the initialized `Manager` instance.
*   **Getter Methods:**
    *   Simple functions that return the corresponding field from the `Manager` struct instance.
        ```go
        // Example Getter
        func (m *Manager) GetNumWorkers() int {
            return m.NumWorkers
        }
        ```

## 6. Error Handling

*   For the MVP, the Configuration Manager itself might not produce many errors if input validation is primarily handled by the CLI component before initialization.
*   If defaults are invalid (developer error), this would be a panic/fatal error at startup.
*   Future enhancements (e.g., loading from a config file) would introduce more error handling (file not found, parse errors).

## 7. Concurrency Considerations

*   The `Manager` struct is populated once at startup and then only read from.
*   Getter methods are read-only operations.
*   Therefore, for the MVP, no explicit locking is required for accessing configuration values after initialization, assuming the `Manager` instance is not modified post-init.

## 8. Dependencies

*   Go standard library: `time`.
*   Internal: `CLIFlags` struct (or equivalent) from the CLI component for initialization values.

## 9. Design Choices & Reasoning

*   **Simplicity for MVP:** Combining hardcoded defaults with CLI overrides is the simplest approach for an MVP. It avoids the complexity of file parsing or environment variable handling initially.
*   **Centralized Access:** Provides a single source of truth for configuration, making it easier to manage and modify settings as the application evolves.
*   **Type Safety:** Storing configuration in a typed struct helps prevent errors.
*   **Extensibility:** While simple, this structure can be extended in Phase 2 to load from environment variables or a configuration file by modifying the `NewManager` logic.

This LLD for the Configuration Manager ensures that all parts of the application can access necessary settings in a consistent and controlled manner.
