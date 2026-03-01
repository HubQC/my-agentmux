package platform

import (
	"os"
	"runtime"
	"strings"
)

// Info holds platform detection results.
type Info struct {
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	IsWSL    bool   `json:"is_wsl"`
	IsDocker bool   `json:"is_docker"`
}

// Detect returns information about the current platform.
func Detect() Info {
	return Info{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		IsWSL:    isWSL(),
		IsDocker: isDocker(),
	}
}

// isWSL detects if running inside Windows Subsystem for Linux.
func isWSL() bool {
	// Check WSL environment variable (set by WSL2)
	if os.Getenv("WSL_DISTRO_NAME") != "" {
		return true
	}
	if os.Getenv("WSLENV") != "" {
		return true
	}

	// Check /proc/version for Microsoft string
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	version := strings.ToLower(string(data))
	return strings.Contains(version, "microsoft") || strings.Contains(version, "wsl")
}

// isDocker detects if running inside a Docker container.
func isDocker() bool {
	// Check /.dockerenv file
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check cgroup for docker
	data, err := os.ReadFile("/proc/1/cgroup")
	if err != nil {
		return false
	}
	content := string(data)
	return strings.Contains(content, "docker") || strings.Contains(content, "containerd")
}
