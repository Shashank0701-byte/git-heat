// Package git provides functions for interacting with git repositories
// via plumbing commands (subprocess execution of git CLI).
package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// ValidateRepo checks if the given path is inside a valid git repository.
// It returns the absolute path to the repository root, or an error.
func ValidateRepo(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path %q: %w", path, err)
	}

	out, err := RunGit(absPath, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("no git repository found at %s", absPath)
	}

	repoRoot := strings.TrimSpace(out)
	if repoRoot == "" {
		return "", fmt.Errorf("no git repository found at %s", absPath)
	}

	return repoRoot, nil
}

// HasCommits checks whether the repository has at least one commit.
func HasCommits(repoRoot string) error {
	_, err := RunGit(repoRoot, "rev-parse", "HEAD")
	if err != nil {
		return fmt.Errorf("repository has no commit history")
	}
	return nil
}

// GetRepoName extracts the repository name from the directory path.
func GetRepoName(repoRoot string) string {
	return filepath.Base(repoRoot)
}

// RunGit executes a git command in the given directory and returns stdout.
// It returns an error if git is not found or the command fails.
func RunGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Check if git is even installed
		if execErr, ok := err.(*exec.Error); ok {
			return "", fmt.Errorf("git not found in PATH: %w", execErr)
		}
		return "", fmt.Errorf("git %s failed: %s", strings.Join(args, " "), strings.TrimSpace(stderr.String()))
	}

	return stdout.String(), nil
}

// RunGitStreaming executes a git command and returns the raw *exec.Cmd
// so callers can stream stdout line-by-line. Caller is responsible for
// calling cmd.Wait() after reading.
func RunGitStreaming(dir string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	// Check git exists before trying to set up pipes
	if _, err := exec.LookPath("git"); err != nil {
		return nil, fmt.Errorf("git not found in PATH: %w", err)
	}

	return cmd, nil
}
