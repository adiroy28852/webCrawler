# Phased Web Crawler Project (Prescriptive Design)

## Overall Project Phases
- [ ] **Phase 1: MVP (CLI-based with PostgreSQL)** - Focus on a functional, single-node crawler with CLI interaction and PostgreSQL for storage.
- [ ] **Phase 2: Cloud Deployment** - Transition the MVP to a cloud environment, incorporating services like AWS SQS for queuing and exploring cloud-native deployment.
- [ ] **Phase 3: Large Scale** - Design and implement enhancements for large-scale, distributed crawling, addressing advanced challenges.

## Current Focus: Phase 1 - MVP (CLI-based with PostgreSQL)

- [x] 1. **Clarify User Preferences & Confirm Phased Approach:** User confirmed preference for a prescriptive, phased approach, starting with a CLI-based MVP using PostgreSQL.

- [x] 2. **Create Prescriptive High-Level Design (HLD) for Phase 1 MVP:**
    - [x] 2a. Define system architecture: Single-node, CLI-driven application.
    - [x] 2b. Specify core components: CLI Interface, Configuration Manager, URL Manager (Frontier), Fetcher, Parser, Storage (PostgreSQL Adapter).
    - [x] 2c. Detail data flow and interactions between components.
    - [x] 2d. Outline key design choices for MVP: Concurrency model (e.g., goroutines for fetch/parse workers), error handling strategy, basic politeness mechanism, simple duplicate URL detection.
    - [x] 2e. Document reasoning for choices, and briefly mention pros/cons of simple alternatives suitable for an MVP.
    - [x] 2f. Create HLD document: `hld_phase1_mvp.md`.

- [x] 3. **Create Prescriptive Low-Level Design (LLD) for Phase 1 Core Components:**
    - [x] 3a. LLD for CLI Interface: Commands (e.g., start crawl, add seed URL, view status), argument parsing.
    - [x] 3b. LLD for Configuration Manager: How to load and manage settings (e.g., user-agent, crawl delay, DB connection).
    - [x] 3c. LLD for URL Manager (Frontier): Data structures for URL queue (e.g., in-memory slice/channel initially), duplicate URL checking (e.g., map), politeness tracking (e.g., map for domain last-access times).
    - [x] 3d. LLD for Fetcher: HTTP request logic, response handling, basic error retries, User-Agent usage.
    - [x] 3e. LLD for Parser: HTML link extraction logic (e.g., using `golang.org/x/net/html` or `goquery`), URL normalization.
    - [x] 3f. LLD for Storage (PostgreSQL Adapter): Database schema (tables for crawled pages, links, etc.), SQL queries for CRUD operations, connection management.
    - [x] 3g. Document key data structures, function/method signatures, core algorithms, and error handling for each component.
    - [x] 3h. Create LLD documents (e.g., `lld_phase1_cli.md`, `lld_phase1_config.md`, `lld_phase1_urlmanager.md`, `lld_phase1_fetcher.md`, `lld_phase1_parser.md`, `lld_phase1_storage_postgres.md`).

- [x] 4. **Update Step-by-Step Implementation Guide for Phase 1 MVP:**
    - [x] 4a. Align the guide with the prescriptive HLD and LLDs for Phase 1.
    - [x] 4b. Provide clear instructions for setting up PostgreSQL for the project.
    - [x] 4c. Guide the implementation of each component based on its LLD.
    - [x] 4d. Explain how to integrate the components.
    - [x] 4e. Create/update the implementation guide: `implementation_guide_phase1.md`.

- [x] 5. **Outline Transition to Phase 2 & 3 in Phase 1 Documents:**
    - [x] 5a. In `hld_phase1_mvp.md`, briefly discuss how the MVP architecture can evolve towards Phase 2 (Cloud) and Phase 3 (Large Scale).
    - [x] 5b. In `implementation_guide_phase1.md`, suggest potential next steps or considerations for future scaling and cloud deployment.

- [ ] 6. **Compile and Send All Phase 1 Prescriptive Design Documents and Guides:**
    - [ ] 6a. Ensure all HLD, LLDs, and the implementation guide for Phase 1 are complete and consistent.
    - [ ] 6b. Send all documents to the user.

- [ ] 7. **Enter Idle State / Await User Feedback for Phase 2 Planning.**
