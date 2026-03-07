package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Info holds git repository information for a working directory.
type Info struct {
	IsRepo    bool
	RepoName  string // basename of the repo root
	Branch    string // current branch
	RemoteURL string // origin remote URL
	RootDir   string // repo root directory
	IsDirty   bool   // has uncommitted changes
	ShortHash string // short HEAD commit hash
}

// Detect returns git information for the given directory.
// Returns a zero Info (IsRepo=false) if the directory is not inside a git repository.
func Detect(workDir string) Info {
	info := Info{}

	// Check if inside a git repo
	root, err := runGit(workDir, "rev-parse", "--show-toplevel")
	if err != nil {
		return info
	}
	info.IsRepo = true
	info.RootDir = strings.TrimSpace(root)
	info.RepoName = filepath.Base(info.RootDir)

	// Get current branch
	branch, err := runGit(workDir, "rev-parse", "--abbrev-ref", "HEAD")
	if err == nil {
		info.Branch = strings.TrimSpace(branch)
	}

	// Get origin remote URL
	remote, err := runGit(workDir, "config", "--get", "remote.origin.url")
	if err == nil {
		info.RemoteURL = strings.TrimSpace(remote)
	}

	// Get short hash
	hash, err := runGit(workDir, "rev-parse", "--short", "HEAD")
	if err == nil {
		info.ShortHash = strings.TrimSpace(hash)
	}

	// Check dirty status
	status, err := runGit(workDir, "status", "--porcelain")
	if err == nil {
		info.IsDirty = strings.TrimSpace(status) != ""
	}

	return info
}

// RepoNameFromRemote extracts a clean repo name from a remote URL.
// e.g., "https://github.com/user/my-project.git" → "my-project"
// e.g., "git@github.com:user/my-project.git" → "my-project"
func RepoNameFromRemote(remoteURL string) string {
	if remoteURL == "" {
		return ""
	}

	// Handle SSH URLs: git@github.com:user/repo.git
	if strings.Contains(remoteURL, ":") && !strings.Contains(remoteURL, "://") {
		parts := strings.SplitN(remoteURL, ":", 2)
		if len(parts) == 2 {
			remoteURL = parts[1]
		}
	}

	// Get the last path component
	name := filepath.Base(remoteURL)

	// Strip .git suffix
	name = strings.TrimSuffix(name, ".git")

	return name
}

// BranchLabel returns a display-friendly branch label.
// For "main"/"master" returns empty (no need to show default branch).
func BranchLabel(branch string) string {
	if branch == "" || branch == "HEAD" {
		return "(detached)"
	}
	if branch == "main" || branch == "master" {
		return ""
	}
	return branch
}

// FormatStatus returns a concise status string for display.
// e.g., "my-project:feature-x*" (with dirty indicator)
func FormatStatus(info Info) string {
	if !info.IsRepo {
		return ""
	}

	parts := []string{info.RepoName}

	label := BranchLabel(info.Branch)
	if label != "" {
		parts = append(parts, ":"+label)
	}

	result := strings.Join(parts, "")
	if info.IsDirty {
		result += "*"
	}
	return result
}

// Checkout switches to the specified branch in the given working directory.
func Checkout(workDir, branch string) error {
	_, err := runGit(workDir, "checkout", branch)
	if err != nil {
		return fmt.Errorf("git checkout %s: %w", branch, err)
	}
	return nil
}

func runGit(workDir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
