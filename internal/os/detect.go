package osutil

import (
	"fmt"
	"runtime"
)

// Detect returns the appropriate Platform for the current OS.
// Currently returns a fully functional Linux implementation.
// Windows and macOS return stubs that will be implemented later.
func Detect() (Platform, error) {
	switch runtime.GOOS {
	case "linux":
		return NewLinuxPlatform()
	case "windows":
		return nil, fmt.Errorf("Windows platform not yet implemented (design only)")
	case "darwin":
		return nil, fmt.Errorf("macOS platform not yet implemented (design only)")
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}
