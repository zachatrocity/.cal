package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/zach/dotcal/internal/calendar"
	"github.com/zach/dotcal/internal/generator"
	"github.com/zach/dotcal/internal/git"
	"github.com/zach/dotcal/internal/logger"
)

type Config struct {
	GithubRepo     string   `json:"githubRepo"`
	GithubBranch   string   `json:"githubBranch"`
	ICSFeeds       []string `json:"icsFeeds"`
	TimeZone       string   `json:"timezone"`
	SyncSchedule   string   `json:"syncSchedule"`
	RepoDirectory  string   `json:"repoDirectory"`
	ScheduleMonths int      `json:"scheduleMonths"`
}

func main() {
	logger.Debug("Starting dotcal application")

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		logger.Error("Failed to load configuration: %v", err)
		os.Exit(1)
	}
	logger.Debug("Configuration loaded: repo=%s, branch=%s, timezone=%s, feeds=%d",
		config.GithubRepo, config.GithubBranch, config.TimeZone, len(config.ICSFeeds))

	// Initialize components
	logger.Debug("Initializing components")
	tz, err := time.LoadLocation(config.TimeZone)
	if err != nil {
		logger.Error("Failed to load timezone: %v", err)
		os.Exit(1)
	}

	fetcher := calendar.NewFetcher()
	parser := calendar.NewParser(tz)
	merger := calendar.NewMerger(tz)
	templateDir := filepath.Join("internal", "templates")
	gen, err := generator.NewGenerator(templateDir)
	if err != nil {
		logger.Error("Failed to initialize generator: %v", err)
		os.Exit(1)
	}
	repo := git.NewRepository(config.RepoDirectory, config.GithubBranch)

	// Ensure we have a valid git repository
	if !repo.IsValidRepo() {
		logger.Debug("No valid git repository found, attempting to clone")
		if err := repo.Clone(config.GithubRepo); err != nil {
			logger.Error("Failed to clone repository: %v", err)
			os.Exit(1)
		}
	}
	logger.Debug("Confirmed valid git repository at %s", config.RepoDirectory)

	// Process calendars
	logger.Debug("Processing calendar feeds")
	var allEvents []calendar.Event
	for _, feedURL := range config.ICSFeeds {
		logger.Debug("Processing feed: %s", feedURL)
		feed := calendar.Feed{
			Source:   feedURL,
			IsURL:    strings.HasPrefix(feedURL, "http"),
			TimeZone: tz,
		}

		logger.Debug("Fetching feed data")
		data, err := fetcher.Fetch(feed)
		if err != nil {
			logger.Error("Failed to fetch feed %s: %v", feedURL, err)
			continue
		}

		events, err := parser.Parse(data)
		if err != nil {
			logger.Error("Failed to parse feed %s: %v", feedURL, err)
			continue
		}

		allEvents = append(allEvents, events...)
	}

	// Generate schedules for configured time range
	logger.Debug("Generating schedules")
	now := time.Now().In(tz)
	startDate := now.AddDate(0, -1, 0) // Start from 1 month ago
	endDate := now.AddDate(0, config.ScheduleMonths, 0)
	logger.Debug("Date range: %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// Track which files we write for commit message
	var updatedFiles []string

	// Generate schedules for each week in the range
	logger.Debug("Processing weeks in range")
	for d := startDate; d.Before(endDate); {
		year, week := d.ISOWeek()

		// Generate schedule for this week
		schedule := merger.MergeEvents(allEvents, year, week)
		content, err := gen.GenerateWeekSchedule(schedule)
		if err != nil {
			logger.Error("Failed to generate schedule for week %d-%d: %v", year, week, err)
			os.Exit(1)
		}

		var filePath string
		if d.Before(now) {
			filePath = fmt.Sprintf("past/%d-W%02d.md", year, week)
		} else {
			filePath = fmt.Sprintf("future/%d-W%02d.md", year, week)
		}

		if err := repo.WriteFile(filePath, content); err != nil {
			logger.Error("Failed to write schedule file %s: %v", filePath, err)
			os.Exit(1)
		}
		updatedFiles = append(updatedFiles, filePath)

		// Move to next week, being careful not to skip partial weeks
		nextDay := d.AddDate(0, 0, 1)
		for nextDay.Weekday() != time.Monday && nextDay.Before(endDate) {
			nextDay = nextDay.AddDate(0, 0, 1)
		}
		d = nextDay
	}

	// Update README.md with current week's schedule
	currentYear, currentWeek := now.ISOWeek()
	var currentWeekPath string
	if now.Weekday() == time.Friday && now.Hour() >= 18 {
		// On Friday evening, use next week's schedule
		nextWeek := now.AddDate(0, 0, 7)
		nextYear, nextWeekNum := nextWeek.ISOWeek()
		currentWeekPath = fmt.Sprintf("future/%d-W%02d.md", nextYear, nextWeekNum)
	} else {
		// Use current week's schedule
		weekStart := calendar.FirstDayOfISOWeek(currentYear, currentWeek, tz)
		if weekStart.Before(now) {
			currentWeekPath = fmt.Sprintf("past/%d-W%02d.md", currentYear, currentWeek)
		} else {
			currentWeekPath = fmt.Sprintf("future/%d-W%02d.md", currentYear, currentWeek)
		}
	}

	// Try both past and future directories if file not found
	currentWeekContent, err := os.ReadFile(filepath.Join(config.RepoDirectory, currentWeekPath))
	if err != nil {
		// If file not found, try the opposite directory
		altPath := currentWeekPath
		if strings.Contains(currentWeekPath, "past/") {
			altPath = strings.Replace(currentWeekPath, "past/", "future/", 1)
		} else {
			altPath = strings.Replace(currentWeekPath, "future/", "past/", 1)
		}

		currentWeekContent, err = os.ReadFile(filepath.Join(config.RepoDirectory, altPath))
		if err != nil {
			logger.Error("Failed to read current week file from both past and future directories: %v", err)
			os.Exit(1)
		}
		// Update the path to where we found the file
		currentWeekPath = altPath
	}

	if err := repo.WriteFile("README.md", string(currentWeekContent)); err != nil {
		logger.Error("Failed to write README.md: %v", err)
		os.Exit(1)
	}
	updatedFiles = append(updatedFiles, "README.md")

	// Commit and push changes
	logger.Debug("Committing changes to repository")
	commitMsg := fmt.Sprintf("Update schedules: %s", strings.Join(updatedFiles, ", "))
	if err := repo.Commit(commitMsg); err != nil {
		logger.Error("Failed to commit changes: %v", err)
		os.Exit(1)
	}

	// Skip push when running under test
	if !testing.Testing() {
		logger.Debug("Pushing changes to remote")
		if err := repo.Push(); err != nil {
			logger.Error("Failed to push changes: %v", err)
			os.Exit(1)
		}
	} else {
		logger.Debug("Skipping push in test mode")
	}

	logger.Info("Successfully updated schedules: %s", strings.Join(updatedFiles, ", "))
	logger.Debug("dotcal application completed successfully")
}

func loadConfig() (*Config, error) {
	// Required environment variables
	githubRepo := os.Getenv("GITHUB_REPO")
	if githubRepo == "" {
		return nil, fmt.Errorf("GITHUB_REPO environment variable is required")
	}

	icsFeeds := os.Getenv("ICS_FEEDS")
	if icsFeeds == "" {
		return nil, fmt.Errorf("ICS_FEEDS environment variable is required")
	}

	config := &Config{
		GithubRepo:     githubRepo,
		GithubBranch:   "main",
		ICSFeeds:       strings.Split(icsFeeds, ","),
		TimeZone:       "UTC",
		SyncSchedule:   "*/30 * * * *",
		RepoDirectory:  "/app/repo",
		ScheduleMonths: 3, // Default to 3 months
	}

	// Load optional environment variables
	if branch := os.Getenv("GITHUB_BRANCH"); branch != "" {
		config.GithubBranch = branch
	}

	if tz := os.Getenv("TIMEZONE"); tz != "" {
		config.TimeZone = tz
	}

	if schedule := os.Getenv("SYNC_SCHEDULE"); schedule != "" {
		config.SyncSchedule = schedule
	}

	if dir := os.Getenv("REPO_DIRECTORY"); dir != "" {
		config.RepoDirectory = dir
	}

	if months := os.Getenv("SCHEDULE_MONTHS"); months != "" {
		if m, err := strconv.Atoi(months); err == nil && m > 0 {
			config.ScheduleMonths = m
		}
	}

	return config, nil
}
