package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	fmt.Println("TaaNOS Icon Setup Tool")
	fmt.Println("----------------------")

	// The source image path (user will place icon.png in TaaNOS root)
	srcPath := "icon.png"

	fmt.Println("Reading user image:", srcPath)
	input, err := os.ReadFile(srcPath)
	if err != nil {
		fmt.Println("Error: Could not find 'icon.png' in the current directory!")
		fmt.Println("Please make sure you put the PNG file in the TaaNOS folder and name it 'icon.png'.")
		return
	}

	fmt.Println("Installing go-winres...")
	cmd := exec.Command("go", "install", "github.com/tc-hib/go-winres@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("Failed to install go-winres:", err)
		return
	}

	if _, err := os.Stat("winres"); os.IsNotExist(err) {
		fmt.Println("Initializing winres directory...")
		cmd := exec.Command("go-winres", "init")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println("Failed to initialize go-winres:", err)
			return
		}
	}

	fmt.Println("Copying icon to winres directory...")
	os.MkdirAll("winres", 0755)
	os.WriteFile(filepath.Join("winres", "icon.png"), input, 0644)

	fmt.Println("Generating Windows resource file...")
	cmd = exec.Command("go-winres", "make")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("Failed to run go-winres make:", err)
		return
	}

	// Move the generated resource.syso to cmd/taanos
	fmt.Println("Moving resource.syso to cmd/taanos...")
	sysoData, err := os.ReadFile("rsrc_windows_amd64.syso")
	if err == nil {
		os.MkdirAll(filepath.Join("cmd", "taanos"), 0755)
		os.WriteFile(filepath.Join("cmd", "taanos", "resource.syso"), sysoData, 0644)
		// Clean up the files generated in root
		files, _ := filepath.Glob("rsrc_windows_*.syso")
		for _, f := range files {
			os.Remove(f)
		}
	} else {
		// Try without arch suffix
		sysoData, err = os.ReadFile("resource.syso")
		if err == nil {
			os.MkdirAll(filepath.Join("cmd", "taanos"), 0755)
			os.WriteFile(filepath.Join("cmd", "taanos", "resource.syso"), sysoData, 0644)
			os.Remove("resource.syso")
		}
	}

	fmt.Println("\n✅ Success! The icon has been embedded.")
	fmt.Println("Now you can run '.\\build.ps1' to compile taanos.exe with the new icon!")
}
