package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/taasezer/TaaNOS/config"
	"github.com/taasezer/TaaNOS/internal/logger"
	"github.com/taasezer/TaaNOS/internal/pipeline"
	"github.com/taasezer/TaaNOS/internal/setup"
	"github.com/taasezer/TaaNOS/internal/tui"
)

const version = config.Version

var (
	newVersionNotice string
	updateCheckDone  = make(chan struct{})
)

func checkUpdateInBackground() {
	defer close(updateCheckDone)

	client := &http.Client{Timeout: 1500 * time.Millisecond}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/taasezer/TaaNOS/releases/latest", nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "TaaNOS-Client")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return
	}
	defer resp.Body.Close()

	var rel struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return
	}

	if isNewerVersion(rel.TagName, version) {
		newVersionNotice = fmt.Sprintf("✨ Yeni TaaNOS sürümü mevcut: %s (Mevcut: v%s). Güncellemek için: taanos upgrade", rel.TagName, version)
	}
}

func parseSemver(v string) (int, int, int) {
	v = strings.TrimPrefix(v, "v")
	for i, r := range v {
		if (r < '0' || r > '9') && r != '.' {
			v = v[:i]
			break
		}
	}
	parts := strings.Split(v, ".")
	var maj, min, patch int
	if len(parts) > 0 {
		fmt.Sscanf(parts[0], "%d", &maj)
	}
	if len(parts) > 1 {
		fmt.Sscanf(parts[1], "%d", &min)
	}
	if len(parts) > 2 {
		fmt.Sscanf(parts[2], "%d", &patch)
	}
	return maj, min, patch
}

func isNewerVersion(latest, current string) bool {
	lMaj, lMin, lPatch := parseSemver(latest)
	cMaj, cMin, cPatch := parseSemver(current)

	if lMaj > cMaj {
		return true
	}
	if lMaj < cMaj {
		return false
	}
	if lMin > cMin {
		return true
	}
	if lMin < cMin {
		return false
	}
	return lPatch > cPatch
}

func printUpdateNoticeIfAny() {
	select {
	case <-updateCheckDone:
	case <-time.After(300 * time.Millisecond):
	}

	if newVersionNotice != "" {
		fmt.Fprintf(os.Stderr, "\n%s\n", newVersionNotice)
	}
}

func main() {
	go checkUpdateInBackground()

	if len(os.Args) < 2 {
		checkFirstRun()
		cmdREPL()
		os.Exit(0)
	}

	command := os.Args[1]

	switch command {
	case "version", "--version", "-V":
		printVersion()

	case "help", "--help", "-h":
		printUsage()

	case "status":
		cmdStatus()

	case "config":
		cmdConfig()

	case "history":
		cmdHistory()

	case "init":
		cmdInit()

	case "model":
		cmdModel()

	case "theme":
		cmdTheme()

	case "uninstall":
		cmdUninstall()

	case "update", "upgrade":
		cmdUpdate()

	default:
		// Everything else is treated as natural language input
		checkFirstRun()
		cmdRun(os.Args[1:])
	}
}

func printVersion() {
	defer printUpdateNoticeIfAny()
	fmt.Printf("TaaNOS v%s\n", version)
	fmt.Println("A deterministic, pipeline-based, local AI-powered CLI system.")
	fmt.Println("https://github.com/taasezer/TaaNOS")
}

func checkFirstRun() {
	if _, err := os.Stat(config.ConfigPath()); os.IsNotExist(err) {
		cfg := config.Default()
		config.AutoDetectModel(cfg)
		_ = config.Save(cfg)
	}
}

func printUsage() {
	defer printUpdateNoticeIfAny()
	fmt.Println(`
╔══════════════════════════════════════════════════════════╗
║                      TaaNOS CLI                          ║
║   Deterministic AI-Powered System Operations Engine      ║
╚══════════════════════════════════════════════════════════╝

USAGE:
  taanos <natural language input>     Execute an operation
  taanos [command]                    Run a subcommand

COMMANDS:
  version       Show version information
  status        Show TaaNOS system status
  config        Show current configuration
  history       Show past chat sessions (history <id> for detail)
  init          First-time setup wizard (Ollama + model detection)
  model         View or change the current AI model
  theme         View or change the TUI theme (taaos, dracula, matrix, ocean, default)
  uninstall     Remove TaaNOS from your system
  update, upgrade Update TaaNOS to the latest version from GitHub

MODE FLAGS:
  -m, --mode    Execution mode: explain | guided | auto  (default: guided)
  -v, --verbose Show detailed output for each pipeline stage
  -n, --dry-run Full pipeline run, but skip actual command execution
  -f, --force   Bypass non-critical validation warnings
  -l, --log-level  Set log verbosity: debug | info | warn | error

EXAMPLES:
  taanos install nginx
  taanos -m explain upgrade all packages
  taanos --dry-run remove docker
  taanos status

SECURITY:
  AI is used ONLY for intent extraction. All commands are deterministic
  and mapped from a hardcoded action registry. No AI-generated commands.`)
	fmt.Println()
}

func cmdREPL() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "taanos: config error: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.New(cfg.Logging.Directory, logger.Level(cfg.Logging.Level))
	if err != nil {
		fmt.Fprintf(os.Stderr, "taanos: logger error: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	if cfg.Theme != "" {
		if t, ok := tui.Themes[cfg.Theme]; ok {
			tui.UpdateStyles(t)
		}
	}

	select {
	case <-updateCheckDone:
	case <-time.After(200 * time.Millisecond):
	}
	if newVersionNotice != "" {
		fmt.Fprintf(os.Stderr, "%s\n", newVersionNotice)
		time.Sleep(1 * time.Second)
	}

	m := tui.New(cfg, log)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "taanos: TUI error: %v\n", err)
		os.Exit(1)
	}
}

func cmdRun(args []string) {
	defer printUpdateNoticeIfAny()
	// Parse flags from args
	mode := pipeline.ModeGuided
	verbose := false
	dryRun := false
	force := false
	var textParts []string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-m", "--mode":
			if i+1 < len(args) {
				i++
				switch args[i] {
				case "explain":
					mode = pipeline.ModeExplain
				case "guided":
					mode = pipeline.ModeGuided
				case "auto":
					mode = pipeline.ModeAuto
				default:
					fmt.Fprintf(os.Stderr, "taanos: unknown mode '%s' (use: explain, guided, auto)\n", args[i])
					os.Exit(1)
				}
			}
		case "-v", "--verbose":
			verbose = true
		case "-n", "--dry-run":
			dryRun = true
		case "-f", "--force":
			force = true
		default:
			textParts = append(textParts, args[i])
		}
	}

	rawText := strings.Join(textParts, " ")
	if rawText == "" {
		fmt.Fprintln(os.Stderr, "taanos: no input provided")
		os.Exit(1)
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "taanos: config error: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.New(cfg.Logging.Directory, logger.Level(cfg.Logging.Level))
	if err != nil {
		fmt.Fprintf(os.Stderr, "taanos: logger error: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	// Build input
	input := pipeline.RawInput{
		RawText:       rawText,
		ExecutionMode: mode,
		Verbose:       verbose,
		DryRun:        dryRun,
		Force:         force,
		Timestamp:     time.Now(),
	}

	// Run pipeline
	p := pipeline.New(cfg, log)
	if err := p.Run(input); err != nil {
		fmt.Fprintf(os.Stderr, "\ntaanos: pipeline error: %v\n", err)
		os.Exit(1)
	}
}

func cmdStatus() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "taanos: config error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("╔══════════════════════════════════╗")
	fmt.Println("║         TaaNOS Status            ║")
	fmt.Println("╠══════════════════════════════════╣")
	fmt.Printf("║ Version:    %-20s ║\n", version)
	fmt.Printf("║ Mode:       %-20s ║\n", cfg.Execution.DefaultMode)
	fmt.Printf("║ Ollama:     %-20s ║\n", cfg.Ollama.Endpoint)
	fmt.Printf("║ Model:      %-20s ║\n", cfg.Ollama.Model)
	fmt.Printf("║ Log Level:  %-20s ║\n", cfg.Logging.Level)
	fmt.Printf("║ Config:     %-20s ║\n", config.ConfigPath())
	fmt.Println("╚══════════════════════════════════╝")
}

func cmdConfig() {
	if len(os.Args) > 2 && os.Args[2] == "set" {
		if len(os.Args) < 5 {
			fmt.Fprintln(os.Stderr, "taanos: usage: taanos config set <key> <value>")
			os.Exit(1)
		}
		fmt.Printf("taanos: config set not yet implemented (Phase 2 skeleton)\n")
		fmt.Printf("  key:   %s\n", os.Args[3])
		fmt.Printf("  value: %s\n", os.Args[4])
		return
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "taanos: config error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Config file: %s\n\n", config.ConfigPath())
	fmt.Printf("ollama:\n")
	fmt.Printf("  endpoint:  %s\n", cfg.Ollama.Endpoint)
	fmt.Printf("  model:     %s\n", cfg.Ollama.Model)
	fmt.Printf("  timeout:   %s\n", cfg.Ollama.Timeout)
	fmt.Printf("  retries:   %d\n", cfg.Ollama.MaxRetries)
	fmt.Printf("\nexecution:\n")
	fmt.Printf("  mode:      %s\n", cfg.Execution.DefaultMode)
	fmt.Printf("  risk_gate: %d\n", cfg.Execution.RequireApprovalAboveRisk)
	fmt.Printf("\nlogging:\n")
	fmt.Printf("  level:     %s\n", cfg.Logging.Level)
	fmt.Printf("  directory: %s\n", cfg.Logging.Directory)
	fmt.Printf("\nsafety:\n")
	fmt.Printf("  max_risk:  %d\n", cfg.Safety.MaxRiskScore)
	fmt.Printf("  root_confirm: %v\n", cfg.Safety.RequireRootConfirmation)
}

func cmdHistory() {
	sessions, err := tui.LoadSessions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "taanos: history error: %v\n", err)
		os.Exit(1)
	}

	if len(sessions) == 0 {
		fmt.Println("No chat sessions yet. Start one with: taanos")
		return
	}

	// If a session ID is provided, show that session's detail
	if len(os.Args) > 2 {
		sessionID := os.Args[2]
		for _, s := range sessions {
			if s.ID == sessionID {
				fmt.Printf("\n╔══════════════════════════════════════════════════════════╗\n")
				fmt.Printf("║  Session: %-47s ║\n", s.StartedAt)
				fmt.Printf("╠══════════════════════════════════════════════════════════╣\n")
				for _, h := range s.History {
					icon := "💬"
					if h.IsPipeline {
						icon = "⚙️ "
					}
					if h.IsErr {
						icon = "❌"
					}
					fmt.Printf("║ %s %-52s ║\n", icon, truncate(h.Input, 50))
				}
				fmt.Printf("╠══════════════════════════════════════════════════════════╣\n")
				fmt.Printf("║  Resume: taanos  (session memory auto-loaded)            ║\n")
				fmt.Printf("╚══════════════════════════════════════════════════════════╝\n")
				return
			}
		}
		fmt.Fprintf(os.Stderr, "taanos: session '%s' not found\n", sessionID)
		os.Exit(1)
	}

	// List all sessions
	fmt.Printf("\n╔══════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║              TaaNOS Chat Sessions (%-3d)                 ║\n", len(sessions))
	fmt.Printf("╠══════════════════════════════════════════════════════════╣\n")

	for i := len(sessions) - 1; i >= 0; i-- {
		s := sessions[i]
		msgCount := len(s.History)
		lastMsg := ""
		if msgCount > 0 {
			lastMsg = truncate(s.History[msgCount-1].Input, 30)
		}
		fmt.Printf("║  %-14s  %s  %2d msgs  %-18s ║\n",
			s.ID, s.StartedAt[:16], msgCount, lastMsg)
	}

	fmt.Printf("╠══════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║  Detail:  taanos history <session_id>                    ║\n")
	fmt.Printf("║  Resume:  taanos  (last session auto-loaded)             ║\n")
	fmt.Printf("╚══════════════════════════════════════════════════════════╝\n")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func cmdInit() {
	wiz := setup.NewWizard()
	if err := wiz.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "\ntaanos: setup error: %v\n", err)
		os.Exit(1)
	}
}

func cmdModel() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "taanos: config error: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		fmt.Printf("Current model: %s\n", cfg.Ollama.Model)
		client := &http.Client{Timeout: 3 * time.Second}
		if resp, err := client.Get(cfg.Ollama.Endpoint + "/api/tags"); err == nil {
			var data struct {
				Models []struct {
					Name string `json:"name"`
				} `json:"models"`
			}
			if json.NewDecoder(resp.Body).Decode(&data) == nil && len(data.Models) > 0 {
				fmt.Println("\nInstalled Ollama Models:")
				for _, m := range data.Models {
					if m.Name == cfg.Ollama.Model {
						fmt.Printf("  ● %s (active)\n", m.Name)
					} else {
						fmt.Printf("  ○ %s\n", m.Name)
					}
				}
			}
			resp.Body.Close()
		}
		fmt.Println("\nUsage: taanos model <model_name>")
		return
	}

	newModel := os.Args[2]
	cfg.Ollama.Model = newModel

	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "taanos: failed to save config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Model changed to: %s\n", newModel)
}

func cmdUninstall() {
	fmt.Println("🗑️ Uninstalling TaaNOS...")
	// Logic to remove TaaNOS
	fmt.Println("Please run uninstall.ps1 (Windows) or uninstall.sh (Linux/MacOS) manually from the installation folder.")
	fmt.Println("This feature will be fully automated in v1.5.")
}

func cmdTheme() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "taanos: config error: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		fmt.Printf("Current theme: %s\n", cfg.Theme)
		fmt.Println("Available themes: taaos, default, ocean, dracula, matrix")
		fmt.Println("Usage: taanos theme <theme_name>")
		return
	}

	selected := strings.ToLower(os.Args[2])
	if _, ok := tui.Themes[selected]; !ok {
		fmt.Fprintf(os.Stderr, "taanos: unknown theme '%s'. Available: taaos, default, ocean, dracula, matrix\n", selected)
		os.Exit(1)
	}

	cfg.Theme = selected
	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "taanos: failed to save config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Theme changed to: %s\n", selected)
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func cmdUpdate() {
	force := false
	for _, arg := range os.Args[2:] {
		if arg == "-f" || arg == "--force" {
			force = true
		}
	}

	execPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "taanos: cannot determine executable path: %v\n", err)
		return
	}

	fmt.Println("🔄 TaaNOS Güncelleme Başlatılıyor...")
	fmt.Printf("📍 Mevcut konum: %s (v%s)\n", execPath, version)

	client := &http.Client{Timeout: 30 * time.Second}

	// 1. Try to get release info from GitHub
	var downloadURLs []string

	req, err := http.NewRequest("GET", "https://api.github.com/repos/taasezer/TaaNOS/releases/latest", nil)
	if err == nil {
		req.Header.Set("User-Agent", "TaaNOS-Update-Agent")
		if resp, err := client.Do(req); err == nil && resp.StatusCode == http.StatusOK {
			var rel githubRelease
			if json.NewDecoder(resp.Body).Decode(&rel) == nil {
				fmt.Printf("🌐 GitHub en son sürüm: %s\n", rel.TagName)
				if !force && !isNewerVersion(rel.TagName, version) {
					fmt.Println("✅ Zaten en güncel TaaNOS sürümünü kullanıyorsunuz!")
					fmt.Println("Yeniden indirmeyi zorlamak için: taanos update --force")
					resp.Body.Close()
					return
				}
				for _, a := range rel.Assets {
					if runtime.GOOS == "windows" && strings.HasSuffix(a.Name, ".exe") {
						downloadURLs = append(downloadURLs, a.BrowserDownloadURL)
					} else if runtime.GOOS == "linux" && (a.Name == "taanos-linux" || a.Name == "taanos-linux-amd64") {
						downloadURLs = append(downloadURLs, a.BrowserDownloadURL)
					} else if runtime.GOOS == "darwin" && strings.Contains(a.Name, "darwin") {
						downloadURLs = append(downloadURLs, a.BrowserDownloadURL)
					}
				}
			}
			resp.Body.Close()
		}
	}

	// 2. Add raw GitHub fallback URLs
	if runtime.GOOS == "windows" {
		downloadURLs = append(downloadURLs,
			"https://raw.githubusercontent.com/taasezer/TaaNOS/main/bin/taanos.exe",
			"https://raw.githubusercontent.com/taasezer/TaaNOS/main/bin/TaaNOS-Setup-Windows.exe",
		)
	} else if runtime.GOOS == "linux" {
		downloadURLs = append(downloadURLs,
			"https://raw.githubusercontent.com/taasezer/TaaNOS/main/bin/taanos-linux-amd64",
			"https://raw.githubusercontent.com/taasezer/TaaNOS/main/bin/TaaNOS-Setup-Linux",
		)
	} else if runtime.GOOS == "darwin" {
		if runtime.GOARCH == "arm64" {
			downloadURLs = append(downloadURLs,
				"https://raw.githubusercontent.com/taasezer/TaaNOS/main/bin/taanos-darwin-arm64",
			)
		} else {
			downloadURLs = append(downloadURLs,
				"https://raw.githubusercontent.com/taasezer/TaaNOS/main/bin/taanos-darwin-amd64",
			)
		}
	}

	tmpFile := execPath + ".new"
	downloaded := false

	for _, url := range downloadURLs {
		fmt.Printf("📥 İndiriliyor: %s ...\n", url)
		dlReq, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}
		dlReq.Header.Set("User-Agent", "TaaNOS-Update-Agent")
		dlResp, err := client.Do(dlReq)
		if err != nil || dlResp.StatusCode != http.StatusOK {
			if dlResp != nil {
				dlResp.Body.Close()
			}
			continue
		}

		out, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			dlResp.Body.Close()
			fmt.Printf("⚠️ Yazma izni reddedildi: %v (Yönetici yetkisi ile deneyin)\n", err)
			return
		}
		written, err := io.Copy(out, dlResp.Body)
		out.Close()
		dlResp.Body.Close()

		if err == nil && written > 1024*1024 { // Minimum 1MB to avoid 404 HTML
			downloaded = true
			break
		}
		_ = os.Remove(tmpFile)
	}

	// 3. If remote download failed but we have local Go compiler and source repo
	if !downloaded {
		if _, err := exec.LookPath("go"); err == nil {
			wd, _ := os.Getwd()
			if strings.Contains(strings.ToLower(wd), "taanos") {
				fmt.Println("⚠️ GitHub ikili dosyasına erişilemedi. Yerel Go derleyicisi ile yeniden derleniyor...")
				cmd := exec.Command("go", "install", "./cmd/taanos")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err == nil {
					fmt.Println("✅ TaaNOS başarıyla yerel kaynak koddan yeniden derlendi ve güncellendi!")
					return
				}
			}
		}
		fmt.Println("❌ Güncelleme indirilemedi. Lütfen internet bağlantınızı kontrol edin veya GitHub reposundan güncel sürümü indirin.")
		return
	}

	// 4. Delete old version and replace with new version
	fmt.Println("🗑️ Eski sürüm kaldırılıyor ve yeni sürüm yerine yerleştiriliyor...")
	oldFile := execPath + ".old"
	_ = os.Remove(oldFile)

	if runtime.GOOS == "windows" {
		if err := os.Rename(execPath, oldFile); err != nil {
			fmt.Fprintf(os.Stderr, "taanos: mevcut dosya değiştirilemedi: %v\nLütfen PowerShell'i Yönetici Olarak çalıştırın.\n", err)
			_ = os.Remove(tmpFile)
			return
		}
		if err := os.Rename(tmpFile, execPath); err != nil {
			_ = os.Rename(oldFile, execPath)
			fmt.Fprintf(os.Stderr, "taanos: yeni dosya yazılamadı: %v\n", err)
			return
		}
		_ = os.Remove(oldFile)
	} else {
		_ = os.Chmod(tmpFile, 0755)
		_ = os.Remove(execPath)
		if err := os.Rename(tmpFile, execPath); err != nil {
			fmt.Fprintf(os.Stderr, "taanos: güncelleme yazılamadı: %v (sudo ile deneyin)\n", err)
			_ = os.Remove(tmpFile)
			return
		}
	}

	fmt.Println("✅ Eski versiyon silindi ve TaaNOS başarıyla güncellendi!")
	fmt.Printf("🚀 Mevcut versiyon: v%s\n", version)
}
