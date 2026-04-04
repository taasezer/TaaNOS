//go:build !linux && !windows

package setup

func detectKernel() string {
	return ""
}

func detectRAM() (int64, int64) {
	return 0, 0
}

func detectVirtualization() (bool, string) {
	return false, ""
}
