# TaaNOS v1.0.0 Specifications

This document outlines the core features and functional specifications of TaaNOS v1.0.0.

## 1. Core Engine
### 1.1 Intent Extraction Pipeline
- **Local Inference**: Native integration with Ollama for zero-data-leak AI processing.
- **Strict Parameterization**: AI is restricted to extracting `Category`, `Action`, and `Target` parameters.
- **Confidence Gating**: Requests with confidence scores below 50% are automatically rejected to prevent ambiguity.
- **AI Boundary**: Once intent is extracted, the pipeline becomes 100% deterministic with no further AI influence.

### 1.2 Execution Registry
- **Hardcoded Command Mapping**: Commands are derived from a pre-defined, audited registry of shell templates.
- **Multi-OS Support**: Commands are dynamically selected based on the detected operating system and distribution.
- **Rollback Registry**: Every destructive or state-changing command must have a corresponding "inverse" command for safety.

## 2. Safety & Validation (7-Stage Validator)
### 2.1 Blocked Actions
- Ability to configure a blacklist of forbidden actions (e.g., `rm -rf /`).
### 2.2 Privilege Verification
- Automatic detection of `root` requirements and verification of `sudo` availability.
### 2.3 Disk Space Check
- Heuristic-based estimation of disk requirements; blocks execution if free space is insufficient.
### 2.4 Risk Thresholding
- Numeric risk scoring (1-10) for every plan. Operations exceeding the max risk score are blocked unless `--force` is used.
### 2.5 Dependency Resolution
- Verifies that required system binaries (e.g., `apt`, `systemctl`, `docker`) are present before attempting execution.
### 2.6 Conflict Detection
- Semantic verification of current vs. target state (e.g., preventing installation of an already-installed package).
### 2.7 Target State Validation
- Final reachability check for the intended operation target.

## 3. CLI & UX
### 3.1 Subcommands
- `init`: Automated setup wizard for hardware detection and model optimization.
- `status`: Real-time health check of the AI endpoint and system configuration.
- `history`: Structured view of past executions retrieved from SQLite.
- `model`: Hot-swapping mechanism for AI models.
- `config`: Intuitive CLI for managing the `config.yaml` file.
- `version`: Version tracking for binary updates.

### 3.2 Execution Modes
- `Explain`: Detailed "dry-run" walkthrough of the intended plan and safety report.
- `Guided`: Step-by-step confirmation mode (Default).
- `Auto`: Pipeline execution with a single initial confirmation.

## 4. System Integration
### 4.1 Cross-Platform Compatibility
- Full support for Windows (PowerShell-based execution).
- Comprehensive Linux support (Debian/Ubuntu, RHEL/Fedora, Arch Linux, Alpine).
### 4.2 Resource Awareness
- Real-time monitoring of RAM, CPU, and Disk metrics used for planning and safety gating.

## 5. Logging & History
### 5.1 SQLite Persistence
- Use of SQLite with Write-Ahead-Logging (WAL) for persistent, thread-safe history tracking.
- Captures full plan metadata, individual step results, and precise execution durations.
### 5.2 Structured Logging
- JSON-formatted logs for automated consumption and debugging.
- Configurable log levels (Debug, Info, Warn, Error).
