package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/zach/dotcal/internal/calendar"
)

// Mock calendar feed for testing
const mockICSData = `BEGIN:VCALENDAR
BEGIN:VEVENT
DTSTART:20250215T100000Z
DTEND:20250215T110000Z
SUMMARY:Test Event
END:VEVENT
END:VCALENDAR`

// Helper to set up test environment
func setupTestEnv(t *testing.T) (string, func()) {
	// Create temporary directory for repo
	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, "repo")

	// Create templates directory structure in the working directory
	templatesDir := filepath.Join("internal", "templates", "default")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Copy weekly template from source
	templateContent, err := os.ReadFile(filepath.Join("..", "..", "internal", "templates", "default", "weekly.md.tmpl"))
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(templatesDir, "weekly.md.tmpl"), templateContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Set up required environment variables
	env := map[string]string{
		"GITHUB_REPO":     "git@github.com:test/repo.git",
		"ICS_FEEDS":       filepath.Join(tmpDir, "test.ics"),
		"REPO_DIRECTORY":  repoDir,
		"TIMEZONE":        "UTC",
		"SCHEDULE_MONTHS": "1", // Use 1 month for faster tests
	}

	for k, v := range env {
		os.Setenv(k, v)
	}

	// Create test ICS file
	if err := os.WriteFile(env["ICS_FEEDS"], []byte(mockICSData), 0644); err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		for k := range env {
			os.Unsetenv(k)
		}
	}

	return tmpDir, cleanup
}

func TestLoadConfig(t *testing.T) {
	// Helper to reset environment variables
	cleanup := func() {
		vars := []string{
			"GITHUB_REPO",
			"ICS_FEEDS",
			"GITHUB_BRANCH",
			"TIMEZONE",
			"SYNC_SCHEDULE",
			"REPO_DIRECTORY",
			"SCHEDULE_MONTHS",
		}
		for _, v := range vars {
			os.Unsetenv(v)
		}
	}

	t.Run("required environment variables", func(t *testing.T) {
		cleanup()
		defer cleanup()

		// Test missing required variables
		_, err := loadConfig()
		if err == nil {
			t.Error("Expected error for missing required variables")
		}

		// Set only GITHUB_REPO
		os.Setenv("GITHUB_REPO", "git@github.com:user/repo.git")
		_, err = loadConfig()
		if err == nil {
			t.Error("Expected error for missing ICS_FEEDS")
		}

		// Set minimum required variables
		os.Setenv("ICS_FEEDS", "http://example.com/calendar.ics")
		config, err := loadConfig()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify default values
		if config.GithubBranch != "main" {
			t.Errorf("Expected default branch 'main', got %s", config.GithubBranch)
		}
		if config.TimeZone != "UTC" {
			t.Errorf("Expected default timezone 'UTC', got %s", config.TimeZone)
		}
		if config.SyncSchedule != "*/30 * * * *" {
			t.Errorf("Expected default schedule '*/30 * * * *', got %s", config.SyncSchedule)
		}
		if config.RepoDirectory != "/app/repo" {
			t.Errorf("Expected default repo directory '/app/repo', got %s", config.RepoDirectory)
		}
		if config.ScheduleMonths != 3 {
			t.Errorf("Expected default schedule months 3, got %d", config.ScheduleMonths)
		}
	})

	t.Run("optional environment variables", func(t *testing.T) {
		cleanup()
		defer cleanup()

		// Set all variables
		env := map[string]string{
			"GITHUB_REPO":     "git@github.com:user/repo.git",
			"ICS_FEEDS":       "feed1.ics,feed2.ics",
			"GITHUB_BRANCH":   "develop",
			"TIMEZONE":        "America/New_York",
			"SYNC_SCHEDULE":   "0 * * * *",
			"REPO_DIRECTORY":  "/custom/path",
			"SCHEDULE_MONTHS": "6",
		}

		for k, v := range env {
			os.Setenv(k, v)
		}

		config, err := loadConfig()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expected := &Config{
			GithubRepo:     "git@github.com:user/repo.git",
			GithubBranch:   "develop",
			ICSFeeds:       []string{"feed1.ics", "feed2.ics"},
			TimeZone:       "America/New_York",
			SyncSchedule:   "0 * * * *",
			RepoDirectory:  "/custom/path",
			ScheduleMonths: 6,
		}

		if !reflect.DeepEqual(config, expected) {
			t.Errorf("Expected %+v, got %+v", expected, config)
		}
	})

	t.Run("invalid schedule months", func(t *testing.T) {
		cleanup()
		defer cleanup()

		os.Setenv("GITHUB_REPO", "git@github.com:user/repo.git")
		os.Setenv("ICS_FEEDS", "feed1.ics")
		os.Setenv("SCHEDULE_MONTHS", "invalid")

		config, err := loadConfig()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Should use default value for invalid input
		if config.ScheduleMonths != 3 {
			t.Errorf("Expected default schedule months 3 for invalid input, got %d", config.ScheduleMonths)
		}
	})

	t.Run("multiple ICS feeds", func(t *testing.T) {
		cleanup()
		defer cleanup()

		os.Setenv("GITHUB_REPO", "git@github.com:user/repo.git")
		os.Setenv("ICS_FEEDS", "feed1.ics,feed2.ics,feed3.ics")

		config, err := loadConfig()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expected := []string{"feed1.ics", "feed2.ics", "feed3.ics"}
		if !reflect.DeepEqual(config.ICSFeeds, expected) {
			t.Errorf("Expected feeds %v, got %v", expected, config.ICSFeeds)
		}
	})
}

func TestMainIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create a test repository
	repoDir := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Initialize git repository
	if err := exec.Command("git", "init", repoDir).Run(); err != nil {
		t.Fatalf("Failed to initialize git repository: %v", err)
	}

	// Configure git
	gitConfigs := [][]string{
		{"config", "--global", "user.name", "Test User"},
		{"config", "--global", "user.email", "test@example.com"},
	}

	for _, args := range gitConfigs {
		if err := exec.Command("git", args...).Run(); err != nil {
			t.Fatalf("Failed to run git %v: %v", args, err)
		}
	}

	// Create initial commit
	if err := os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("# Test Repository"), 0644); err != nil {
		t.Fatalf("Failed to create README.md: %v", err)
	}

	// Stage and commit
	gitCommands := [][]string{
		{"add", "."},
		{"commit", "-m", "Initial commit"},
	}

	for _, args := range gitCommands {
		cmd := exec.Command("git", args...)
		cmd.Dir = repoDir // Set working directory for git commands
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to run git %v: %v", args, err)
		}
	}

	// Run main testing.Testing() gates push
	main()

	// Verify expected files exist
	expectedFiles := []string{
		"README.md",
		"past",
		"future",
	}

	for _, f := range expectedFiles {
		if _, err := os.Stat(filepath.Join(repoDir, f)); os.IsNotExist(err) {
			t.Errorf("Expected file/directory %s to exist", f)
		}
	}
}

// TestMainComponents tests individual components used in main
func TestMainComponents(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	t.Run("calendar processing", func(t *testing.T) {
		tz := time.UTC
		fetcher := calendar.NewFetcher()
		parser := calendar.NewParser(tz)

		feed := calendar.Feed{
			Source:   filepath.Join(tmpDir, "test.ics"),
			IsURL:    false,
			TimeZone: tz,
		}

		data, err := fetcher.Fetch(feed)
		if err != nil {
			t.Fatalf("Failed to fetch calendar data: %v", err)
		}

		events, err := parser.Parse(data)
		if err != nil {
			t.Fatalf("Failed to parse calendar data: %v", err)
		}

		if len(events) != 1 {
			t.Errorf("Expected 1 event, got %d", len(events))
		}

		if events[0].Title != "Test Event" {
			t.Errorf("Expected event title 'Test Event', got '%s'", events[0].Title)
		}
	})
}
