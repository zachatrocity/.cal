package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Repository handles git operations
type Repository struct {
	path   string
	branch string
}

// NewRepository creates a new git repository handler
func NewRepository(path, branch string) *Repository {
	return &Repository{
		path:   path,
		branch: branch,
	}
}

// Clone clones the repository
func (r *Repository) Clone(url string) error {
	if _, err := os.Stat(r.path); !os.IsNotExist(err) {
		return fmt.Errorf("destination path already exists: %s", r.path)
	}

	cmd := exec.Command("git", "clone", "--branch", r.branch, url, r.path)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}

// Pull updates the repository
func (r *Repository) Pull() error {
	cmd := exec.Command("git", "-C", r.path, "pull", "origin", r.branch)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull repository: %w", err)
	}

	return nil
}

// WriteFile writes content to a file in the repository
func (r *Repository) WriteFile(filename string, content string) error {
	fullPath := filepath.Join(r.path, filename)

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Commit commits changes
func (r *Repository) Commit(message string) error {
	// Add all changes
	cmd := exec.Command("git", "-C", r.path, "add", ".")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	// Check if there are changes to commit
	cmd = exec.Command("git", "-C", r.path, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}

	if len(strings.TrimSpace(string(output))) == 0 {
		return nil // No changes to commit
	}

	// Commit changes
	cmd = exec.Command("git", "-C", r.path, "commit", "-m", message)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	return nil
}

// Push pushes changes to remote
func (r *Repository) Push() error {
	cmd := exec.Command("git", "-C", r.path, "push", "origin", r.branch)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push changes: %w", err)
	}

	return nil
}
