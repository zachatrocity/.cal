package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/zach/dotcal/internal/logger"
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
	logger.Debug("Attempting to clone repository from %s to %s (branch: %s)", url, r.path, r.branch)

	if _, err := os.Stat(r.path); !os.IsNotExist(err) {
		logger.Debug("Destination path already exists: %s", r.path)
		return fmt.Errorf("destination path already exists: %s", r.path)
	}

	cmd := exec.Command("git", "clone", "--branch", r.branch, url, r.path)
	logger.Debug("Running command: git clone --branch %s %s %s", r.branch, url, r.path)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("Clone failed: %v\nOutput: %s", err, string(output))
		return fmt.Errorf("failed to clone repository: %w\nOutput: %s", err, string(output))
	}

	logger.Debug("Repository cloned successfully")
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
	logger.Debug("Attempting to commit changes in %s", r.path)

	// Add all changes
	cmd := exec.Command("git", "-C", r.path, "add", ".")
	logger.Debug("Running command: git -C %s add .", r.path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("Failed to stage changes: %v\nOutput: %s", err, string(output))
		return fmt.Errorf("failed to stage changes: %w\nOutput: %s", err, string(output))
	}

	// Check if there are changes to commit
	cmd = exec.Command("git", "-C", r.path, "status", "--porcelain")
	logger.Debug("Running command: git -C %s status --porcelain", r.path)
	output, err = cmd.Output()
	if err != nil {
		logger.Error("Failed to check git status: %v", err)
		return fmt.Errorf("failed to check git status: %w", err)
	}

	if len(strings.TrimSpace(string(output))) == 0 {
		logger.Debug("No changes to commit")
		return nil // No changes to commit
	}

	logger.Debug("Changes to commit:\n%s", string(output))

	// Commit changes
	cmd = exec.Command("git", "-C", r.path, "commit", "-m", message)
	logger.Debug("Running command: git -C %s commit -m \"%s\"", r.path, message)
	output, err = cmd.CombinedOutput()
	if err != nil {
		logger.Error("Failed to commit changes: %v\nOutput: %s", err, string(output))
		return fmt.Errorf("failed to commit changes: %w\nOutput: %s", err, string(output))
	}

	logger.Debug("Changes committed successfully")
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
