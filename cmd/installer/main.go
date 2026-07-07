package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

//go:embed payload
var payload []byte

func main() {
	fmt.Println("╔════════════════════════════════════════════════╗")
	fmt.Println("║               TaaNOS Installer                 ║")
	fmt.Println("╚════════════════════════════════════════════════╝")

	var installDir string
	var exeName string

	if runtime.GOOS == "windows" {
		installDir = filepath.Join(os.Getenv("ProgramFiles"), "TaaNOS")
		exeName = "taanos.exe"
	} else {
		installDir = "/usr/local/bin"
		exeName = "taanos"
	}

	fmt.Printf("\nInstalling TaaNOS to: %s\n", installDir)

	if runtime.GOOS == "windows" {
		err := os.MkdirAll(installDir, 0755)
		if err != nil {
			fmt.Println("Error creating directory. Please run as Administrator!")
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		// Linux/Mac typically requires sudo for /usr/local/bin
		// We'll just attempt to write. If it fails, they need sudo.
	}

	targetPath := filepath.Join(installDir, exeName)
	err := os.WriteFile(targetPath, payload, 0755)
	if err != nil {
		fmt.Printf("Error writing file (Do you need sudo/Admin?): %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Successfully installed to", targetPath)

	if runtime.GOOS == "windows" {
		// Try to add to Path via powershell
		fmt.Println("Adding to System PATH...")
		cmd := exec.Command("powershell", "-Command", fmt.Sprintf(`[Environment]::SetEnvironmentVariable("Path", $env:Path + ";%s", "User")`, installDir))
		cmd.Run()
		fmt.Println("✅ Path updated. Please restart your terminal.")
	}

	fmt.Println("\nInstallation Complete! You can now run 'taanos' in your terminal.")
}
