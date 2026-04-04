# TaaNOS
**A Deterministic, Pipeline-Based Local AI-Powered CLI System**

TaaNOS is a next-generation command-line interface that allows you to manage any system using natural language. Unlike standard AI assistants that blindly generate scripts, TaaNOS leverages AI strictly for intent extraction and relies on a hardcoded, deterministic execution registry.

This ensures complete safety, transparency, and traceability for system operations.

---

## Features

- **Local AI Only**: Uses Ollama entirely locally. No internet connection required for inference. No data leaves your machine.
- **Deterministic Predictability**: AI is not allowed to generate commands. It only identifies the "Intent". The actual commands come from a deterministic, hardcoded action registry.
- **Cross-Platform Support**: Natively supports both Windows and Linux environments, with support for major package managers (APT, DNF, Pacman, APK).
- **Safety First**: Incorporates a 7-stage validator that blocks risky actions, verifies root privileges, evaluates disk space, and checks for tool dependencies before execution.
- **Auto-Recovery**: If a step fails during execution, TaaNOS automatically attempts to roll back previously completed steps using predefined inverse actions.
- **Execution History**: Built-in SQLite database tracks every operation, capturing intent, results, duration, and risk profiles for full auditability.
- **Explain Mode**: Dry-run capability that shows exactly what would happen without touching the system.

---

## Installation and Setup

### Linux / macOS:
```bash
curl -fsSL https://raw.githubusercontent.com/taasezer/TaaNOS/main/install.sh | bash
```

### Windows (PowerShell):
```powershell
Invoke-Expression (Invoke-WebRequest -Uri "https://raw.githubusercontent.com/taasezer/TaaNOS/main/install.ps1" -UseBasicParsing).Content
```

*(Alternatively, if you have Go installed, you can run: `go install github.com/taasezer/TaaNOS/cmd/taanos@latest`)*

### First-Time Initialization:
The built-in wizard detects your OS, virtualization status, and available RAM to automatically install Ollama and configure the optimal AI model.
```bash
taanos init
```

---

## Usage

Interact with TaaNOS using natural language or dedicated subcommands.

### General Execution:
```bash
taanos install nginx
taanos remove docker completely
taanos start the apache service
```

### Mode Flags:
- `-m, --mode`: Execution mode (`explain`, `guided`, `auto`).
- `-v, --verbose`: Show detailed output for each pipeline stage.
- `-n, --dry-run`: Full pipeline run without actual command execution.
- `-f, --force`: Bypass non-critical validation warnings.
- `-l, --log-level`: Set log verbosity (`debug`, `info`, `warn`, `error`).

### Available Modes:
- **explain**: Explains the plan and safety checks without executing.
- **guided** (Default): Requires user confirmation for each step in the pipeline.
- **auto**: Executes the entire pipeline automatically after one initial confirmation.

---

## Commands

- `taanos init` — Run the setup wizard for Ollama and model configuration.
- `taanos status` — View system status, AI endpoint, model, and configuration path.
- `taanos history` — List recent operations, success rates, and execution details from the SQLite store.
- `taanos model` — View or change the current AI model used for intent extraction.
- `taanos config` — List existing configuration and safety gating levels.
- `taanos version` — Show current version and build information.

---

## Architecture: The 9 Stages of TaaNOS

TaaNOS operates through a structured pipeline to ensure deterministic results from non-deterministic AI inputs:

1. **Input Parsing**: Extracts flags and cleanses natural language input.
2. **Intent Extraction (AI Boundary)**: Uses Ollama to extract Category, Action, and Target parameters against a strict internal schema.
3. **Context Analysis**: Probes the system for OS details, package manager status, user privileges, and available resources.
4. **Planning (Zero-AI)**: Map the intent to a deterministic sequence of steps from the hardcoded execution registry.
5. **Validation**: Runs 7 safety checks including blocked actions, disk space, and risk thresholding.
6. **Interaction**: Presents the plan and validation report to the user for final approval.
7. **Execution**: Translates steps into system processes with stdout/stderr capture and timeout management.
8. **Recovery**: Triggers rollback actions for previous steps if a failure occurs mid-execution.
9. **History Logging**: Commits the entire lifecycle profile to a local SQLite write-ahead-log database.

---

## License
MIT