package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestRepo(t *testing.T) (string, func()) {
	// Create a temporary directory for the test repository
	tmpDir := t.TempDir()

	// Initialize a git repository
	cmd := exec.Command("git", "init", tmpDir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize test repository: %v", err)
	}

	// Set git config for test commits
	cmd = exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git config user.name: %v", err)
	}

	cmd = exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git config user.email: %v", err)
	}

	// Create initial commit
	readme := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readme, []byte("# Test Repository"), 0644); err != nil {
		t.Fatalf("Failed to create README.md: %v", err)
	}

	cmd = exec.Command("git", "-C", tmpDir, "add", "README.md")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to stage README.md: %v", err)
	}

	cmd = exec.Command("git", "-C", tmpDir, "commit", "-m", "Initial commit")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create initial commit: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestNewRepository(t *testing.T) {
	path := "/test/path"
	branch := "main"
	repo := NewRepository(path, branch)

	if repo.path != path {
		t.Errorf("Expected path %s, got %s", path, repo.path)
	}
	if repo.branch != branch {
		t.Errorf("Expected branch %s, got %s", branch, repo.branch)
	}
}

func TestIsValidRepo(t *testing.T) {
	t.Run("valid repository", func(t *testing.T) {
		path, cleanup := setupTestRepo(t)
		defer cleanup()

		repo := NewRepository(path, "main")
		if !repo.IsValidRepo() {
			t.Error("Expected valid repository")
		}
	})

	t.Run("invalid repository", func(t *testing.T) {
		tmpDir := t.TempDir()
		repo := NewRepository(tmpDir, "main")
		if repo.IsValidRepo() {
			t.Error("Expected invalid repository")
		}
	})
}

func TestWriteFile(t *testing.T) {
	path, cleanup := setupTestRepo(t)
	defer cleanup()

	repo := NewRepository(path, "main")
	testFile := "test.txt"
	content := "test content"

	t.Run("write new file", func(t *testing.T) {
		if err := repo.WriteFile(testFile, content); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		// Verify file content
		data, err := os.ReadFile(filepath.Join(path, testFile))
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		if string(data) != content {
			t.Errorf("Expected content %q, got %q", content, string(data))
		}
	})

	t.Run("write file in subdirectory", func(t *testing.T) {
		subPath := "sub/dir/test.txt"
		if err := repo.WriteFile(subPath, content); err != nil {
			t.Fatalf("Failed to write file in subdirectory: %v", err)
		}

		// Verify file content
		data, err := os.ReadFile(filepath.Join(path, subPath))
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		if string(data) != content {
			t.Errorf("Expected content %q, got %q", content, string(data))
		}
	})
}

func TestCommit(t *testing.T) {
	path, cleanup := setupTestRepo(t)
	defer cleanup()

	repo := NewRepository(path, "main")

	t.Run("commit with changes", func(t *testing.T) {
		// Create a new file
		if err := repo.WriteFile("test.txt", "test content"); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		message := "Test commit"
		if err := repo.Commit(message); err != nil {
			t.Fatalf("Failed to commit changes: %v", err)
		}

		// Verify commit
		cmd := exec.Command("git", "-C", path, "log", "-1", "--pretty=format:%s")
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Failed to get commit message: %v", err)
		}
		if string(output) != message {
			t.Errorf("Expected commit message %q, got %q", message, string(output))
		}
	})

	t.Run("commit without changes", func(t *testing.T) {
		if err := repo.Commit("No changes"); err != nil {
			t.Fatalf("Expected no error for commit without changes, got %v", err)
		}
	})
}

func TestInitializeRepoStructure(t *testing.T) {
	path, cleanup := setupTestRepo(t)
	defer cleanup()

	repo := NewRepository(path, "main")
	if err := repo.initializeRepoStructure(); err != nil {
		t.Fatalf("Failed to initialize repository structure: %v", err)
	}

	// Verify directories
	dirs := []string{"past", "future"}
	for _, dir := range dirs {
		dirPath := filepath.Join(path, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			t.Errorf("Expected directory %s to exist", dir)
		}
	}

	// Verify README.md
	readmePath := filepath.Join(path, "README.md")
	content, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("Failed to read README.md: %v", err)
	}
	if !strings.Contains(string(content), "DotCal Schedule") {
		t.Error("README.md does not contain expected content")
	}
}

func TestClone(t *testing.T) {
	// Create a source repository
	sourceDir, sourceCleanup := setupTestRepo(t)
	defer sourceCleanup()

	repo := NewRepository(filepath.Join(t.TempDir(), "repo"), "main")

	t.Run("clone repository", func(t *testing.T) {
		if err := repo.Clone("file://" + sourceDir); err != nil {
			t.Fatalf("Failed to clone repository: %v", err)
		}

		// Verify clone
		if !repo.IsValidRepo() {
			t.Error("Expected valid repository after clone")
		}

		// Verify branch
		cmd := exec.Command("git", "-C", repo.path, "rev-parse", "--abbrev-ref", "HEAD")
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Failed to get current branch: %v", err)
		}
		if strings.TrimSpace(string(output)) != "main" {
			t.Errorf("Expected branch main, got %s", strings.TrimSpace(string(output)))
		}
	})

	t.Run("clone to existing directory", func(t *testing.T) {
		existingDir := filepath.Join(t.TempDir(), "existing")
		if err := os.MkdirAll(existingDir, 0755); err != nil {
			t.Fatalf("Failed to create existing directory: %v", err)
		}
		if err := os.WriteFile(filepath.Join(existingDir, "existing.txt"), []byte("existing"), 0644); err != nil {
			t.Fatalf("Failed to create existing file: %v", err)
		}

		repo := NewRepository(existingDir, "main")
		err := repo.Clone("file://" + sourceDir)
		if err == nil {
			t.Error("Expected error when cloning to existing directory")
		}
	})
}
