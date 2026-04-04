package setup

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/taasezer/TaaNOS/config"
)

// ModelRecommendation defines a model suggestion based on system resources.
type ModelRecommendation struct {
	Name        string
	SizeMB      int
	MinRAMMB    int
	Description string
}

// Available models in priority order (smallest to largest)
var modelTiers = []ModelRecommendation{
	{"tinyllama", 600, 0, "Tiny model — works on any system (< 2GB RAM)"},
	{"phi3:mini", 1500, 2048, "Small model — good for 2-4GB RAM systems"},
	{"llama3.2", 2300, 4096, "Standard model — recommended for 4-8GB RAM"},
	{"llama3.1", 4700, 8192, "Full model — best quality, needs 8GB+ RAM"},
}

// Wizard handles first-run setup: Ollama detection, VM detection, model recommendation.
type Wizard struct {
	reader *bufio.Reader
}

// NewWizard creates a new setup wizard.
func NewWizard() *Wizard {
	return &Wizard{
		reader: bufio.NewReader(os.Stdin),
	}
}

// Run executes the full setup wizard.
func (w *Wizard) Run() error {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║              TaaNOS — First-Time Setup                  ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Step 1: System detection
	sysInfo := w.detectSystem()
	w.displaySystemInfo(sysInfo)

	// Step 2: Ollama detection
	ollamaInstalled := w.checkOllama()
	if !ollamaInstalled {
		installed := w.offerOllamaInstall()
		if !installed {
			fmt.Println("\n⚠️  Ollama is required for TaaNOS. Install it manually:")
			fmt.Println("   curl -fsSL https://ollama.com/install.sh | sh")
			return fmt.Errorf("Ollama not installed")
		}
	}

	// Step 3: Model recommendation
	recommended := w.recommendModel(sysInfo.AvailableRAMMB)
	model := w.offerModelSelection(recommended, sysInfo.AvailableRAMMB)

	// Step 4: Pull model
	if model != "" {
		w.pullModel(model)
	}

	// Step 5: Configure
	w.configureSystem(model, sysInfo)

	fmt.Println("\n╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║              ✅ TaaNOS Setup Complete!                   ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Printf("\n  Try: taanos install nginx\n")
	fmt.Printf("  Or:  taanos -m explain update packages\n\n")

	return nil
}

// SystemInfo holds detected system information.
type SystemInfo struct {
	OS             string
	Arch           string
	Kernel         string
	IsVM           bool
	VMType         string
	TotalRAMMB     int64
	AvailableRAMMB int64
	DiskFreeMB     int64
}

// detectSystem gathers system information.
func (w *Wizard) detectSystem() SystemInfo {
	fmt.Println("🔍 Detecting system...")

	info := SystemInfo{
		OS:     runtime.GOOS,
		Arch:   runtime.GOARCH,
		Kernel: detectKernel(),
	}

	// Detect VM
	info.IsVM, info.VMType = detectVirtualization()

	// Detect RAM
	info.TotalRAMMB, info.AvailableRAMMB = detectRAM()

	return info
}

// displaySystemInfo shows what was detected.
func (w *Wizard) displaySystemInfo(info SystemInfo) {
	fmt.Printf("\n  System:     %s/%s\n", info.OS, info.Arch)
	if info.Kernel != "" {
		fmt.Printf("  Kernel:     %s\n", info.Kernel)
	}
	if info.IsVM {
		fmt.Printf("  VM:         ✅ Yes (%s)\n", info.VMType)
	} else {
		fmt.Printf("  VM:         ❌ Bare metal\n")
	}
	fmt.Printf("  Total RAM:  %d MB\n", info.TotalRAMMB)
	fmt.Printf("  Free RAM:   %d MB\n", info.AvailableRAMMB)
	fmt.Println()
}

// checkOllama checks if Ollama is installed and running.
func (w *Wizard) checkOllama() bool {
	fmt.Println("🔍 Checking for Ollama...")

	// Check if binary exists
	_, err := exec.LookPath("ollama")
	if err != nil {
		fmt.Println("  ❌ Ollama not found")
		return false
	}
	fmt.Println("  ✅ Ollama binary found")

	// Check if API is reachable
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://localhost:11434/api/version")
	if err != nil {
		fmt.Println("  ⚠️  Ollama installed but not running")
		fmt.Println("     Start it with: ollama serve")
		return true // installed but not serving
	}
	defer resp.Body.Close()

	fmt.Println("  ✅ Ollama API reachable")
	return true
}

// offerOllamaInstall offers to install Ollama automatically.
func (w *Wizard) offerOllamaInstall() bool {
	if runtime.GOOS != "linux" {
		fmt.Println("\n  Ollama auto-install is only available on Linux.")
		fmt.Println("  Download from: https://ollama.com/download")
		return false
	}

	fmt.Print("\n  Install Ollama now? [Y/n]: ")
	response := w.prompt()

	if response == "n" || response == "no" {
		return false
	}

	fmt.Println("\n  📦 Installing Ollama...")
	cmd := exec.Command("bash", "-c", "curl -fsSL https://ollama.com/install.sh | sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("  ❌ Installation failed: %v\n", err)
		return false
	}

	fmt.Println("  ✅ Ollama installed successfully")

	// Start Ollama service
	fmt.Println("  Starting Ollama service...")
	startCmd := exec.Command("bash", "-c", "nohup ollama serve > /dev/null 2>&1 &")
	startCmd.Run()
	time.Sleep(2 * time.Second) // Give it time to start

	return true
}

// recommendModel suggests the best model for the system's RAM.
func (w *Wizard) recommendModel(availableRAMMB int64) *ModelRecommendation {
	var best *ModelRecommendation
	for i := len(modelTiers) - 1; i >= 0; i-- {
		tier := modelTiers[i]
		if availableRAMMB >= int64(tier.MinRAMMB) || i == 0 {
			best = &modelTiers[i]
			break
		}
	}
	return best
}

// offerModelSelection lets the user choose a model.
func (w *Wizard) offerModelSelection(recommended *ModelRecommendation, availableRAM int64) string {
	fmt.Println("🤖 Model Selection")
	fmt.Printf("   Recommended: %s (%s)\n\n", recommended.Name, recommended.Description)

	fmt.Println("   Available models:")
	for i, m := range modelTiers {
		marker := "  "
		if m.Name == recommended.Name {
			marker = "→ "
		}
		ramWarn := ""
		if availableRAM > 0 && availableRAM < int64(m.MinRAMMB) {
			ramWarn = " ⚠️  (needs more RAM)"
		}
		fmt.Printf("   %s%d. %-12s %4d MB  %s%s\n", marker, i+1, m.Name, m.SizeMB, m.Description, ramWarn)
	}

	fmt.Printf("\n   Choose model [1-%d] or press Enter for recommended: ", len(modelTiers))
	response := w.prompt()

	if response == "" {
		return recommended.Name
	}

	idx, err := strconv.Atoi(response)
	if err != nil || idx < 1 || idx > len(modelTiers) {
		fmt.Println("   Using recommended model")
		return recommended.Name
	}

	return modelTiers[idx-1].Name
}

// pullModel downloads the selected model.
func (w *Wizard) pullModel(model string) {
	fmt.Printf("\n📥 Pulling model '%s'...\n", model)
	fmt.Println("   This may take a few minutes depending on your connection.")

	cmd := exec.Command("ollama", "pull", model)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("   ⚠️  Model pull failed: %v\n", err)
		fmt.Printf("   You can pull it manually: ollama pull %s\n", model)
		return
	}

	fmt.Printf("   ✅ Model '%s' ready\n", model)
}

// configureSystem writes the correct config based on detected system.
func (w *Wizard) configureSystem(model string, info SystemInfo) {
	fmt.Println("\n⚙️  Configuring TaaNOS...")

	cfg, _ := config.Load()

	// Set model
	if model != "" {
		cfg.Ollama.Model = model
	}

	// Adjust timeout based on system
	if info.IsVM || info.AvailableRAMMB < 4096 {
		cfg.Ollama.Timeout = 120 * time.Second // More time for slow systems
		fmt.Println("   Timeout: 120s (adjusted for VM/low-RAM)")
	} else {
		cfg.Ollama.Timeout = 30 * time.Second
		fmt.Println("   Timeout: 30s")
	}

	// Save config
	if err := config.Save(cfg); err != nil {
		fmt.Printf("   ⚠️  Failed to save config: %v\n", err)
		return
	}

	fmt.Printf("   Model: %s\n", cfg.Ollama.Model)
	fmt.Printf("   Config saved to: %s\n", config.ConfigPath())
}

// prompt reads user input.
func (w *Wizard) prompt() string {
	input, _ := w.reader.ReadString('\n')
	return strings.TrimSpace(strings.ToLower(input))
}

