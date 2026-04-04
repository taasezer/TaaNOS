# TaaNOS
**A Deterministic, Pipeline-Based Local AI-Powered CLI System**

TaaNOS is a next-generation command-line interface that allows you to manage any system using natural language. Unlike standard AI assistants that blindly generate scripts, TaaNOS leverages AI *strictly* for intent extraction and relies on a hardcoded, deterministic execution registry. 

This ensures complete safety, transparency, and traceability for system operations.

---

## 🚀 Features

- **Local AI Only**: Uses Ollama entirely locally. No internet connection required for inference. No data leaves your machine.
- **Deterministic Predictability**: AI is not allowed to generate commands. It only identifies the "Intent". The actual commands come from a deterministic, hardcoded action registry.
- **Cross-Platform**: Natively supports both Windows and Linux (deb, rpm, pacman, apk) environments.
- **Safety First**: Incorporates a 7-stage validator that blocks risky actions, requires `sudo` validation, and evaluates disk space and dependencies before anything executes.
- **Auto-Recovery**: If a step fails mid-execution, TaaNOS automatically attempts to roll back any previously completed steps.
- **Execution History**: Built-in SQLite history allows you to view past executions, search by intent, and view risk profiles.

---

## 🛠️ Installation & Setup

**For Linux / macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/taasezer/TaaNOS/main/install.sh | bash
```

**For Windows (PowerShell):**
```powershell
Invoke-Expression (Invoke-WebRequest -Uri "https://raw.githubusercontent.com/taasezer/TaaNOS/main/install.ps1" -UseBasicParsing).Content
```

*(Alternatively, if you have Go installed, you can run: `go install github.com/taasezer/TaaNOS/cmd/taanos@latest`)*

1. **First-Time Initialization Wizard**:
   The built-in wizard will detect your OS, Virtual Machine status, and available RAM to automatically install Ollama and allocate the best model for your system.
   ```bash
   taanos init
   ```

## 💻 Usage

Just tell TaaNOS what you want to do.

**General Execution:**
```bash
taanos install nginx
taanos remove docker completely
taanos start the apache service
```

**Explain Mode (Dry Run/Preview):**
```bash
taanos -m explain secure my ssh config
```

**Available Modes (`-m`)**:
- `explain`: Just explains the plan without executing.
- `guided` (Default): Requires you to press `Y/n` at each step of the pipeline.
- `auto`: Executes the entire pipeline automatically after one initial confirmation.

**Commands**:
- `taanos history` — View the last 20 operations you executed, including duration and success rates.
- `taanos status` — View your configuration, model, memory profiling, and AI endpoint status.
- `taanos config` — List out your existing `config.yaml` and safety gating levels.

## 🏗️ Architecture (The 9 Stages of TaaNOS)

TaaNOS is built as a highly structured pipeline spanning 9 discrete stages:

1. **CLI Parsing**: Extracts flags (`-n`, `-v`, `-m`) and parses natural language inputs.
2. **Intent Engine (AI Boundary)**: Asks Ollama to extract the exact `Category`, `Action`, and `Target` from a strict enum list.
3. **Context Analyzer**: Evaluates your OS, Distro, Package Manager, and current User Privilege.
4. **Planner (Zero-AI)**: Looks up the requested Intent in the deterministic `registry.go` and strings together the required steps (ex: `update package repo` → `install target`).
5. **Validator**: Runs 7 deterministic safety checks. Blocks execution if Risk Level > Maximum threshold, or disk space is insufficient.
6. **Interaction**: Determines how to prompt the user (Explain vs Guided vs Auto).
7. **Executor**: Translates steps into system processes, manages timeouts, handles `sudo` escalation, and captures stdout/stderr.
8. **Recovery**: If execution hits a failure status, walks backwards and triggers reverse target actions if available.
9. **History Logging**: Commits the entire lifecycle profile to a SQLite write-ahead-log database.

## 📜 License
MIT