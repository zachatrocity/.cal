package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/zach/dotcal/internal/anonymizer"
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
	logger.Debug("Starting DotCal application")

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
	anon := anonymizer.NewAnonymizer()
	gen := generator.NewGenerator(anon)
	repo := git.NewRepository(config.RepoDirectory, config.GithubBranch)

	// Clone repository if it doesn't exist
	if _, err := os.Stat(config.RepoDirectory); os.IsNotExist(err) {
		if err := repo.Clone(config.GithubRepo); err != nil {
			logger.Error("Failed to clone repository: %v", err)
			os.Exit(1)
		}
	}

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
	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 7) {
		year, week := d.ISOWeek()
		schedule := merger.MergeEvents(allEvents, year, week)
		content := gen.GenerateWeekSchedule(schedule)

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

		// Handle README.md update
		if d.Equal(now.Truncate(24 * time.Hour)) {
			// It's the current week
			if now.Weekday() == time.Friday && now.Hour() >= 18 {
				// Friday evening - use next week's schedule
				nextWeek := d.AddDate(0, 0, 7)
				nextYear, nextWeekNum := nextWeek.ISOWeek()
				nextSchedule := merger.MergeEvents(allEvents, nextYear, nextWeekNum)
				nextContent := gen.GenerateWeekSchedule(nextSchedule)

				// Move current README.md to past
				if err := repo.WriteFile(fmt.Sprintf("past/%d-W%02d.md", year, week), content); err != nil {
					logger.Error("Failed to archive current README.md: %v", err)
					os.Exit(1)
				}

				// Update README.md with next week
				if err := repo.WriteFile("README.md", nextContent); err != nil {
					logger.Error("Failed to update README.md: %v", err)
					os.Exit(1)
				}
				updatedFiles = append(updatedFiles, "README.md")
			} else {
				// Not Friday evening - use current week
				if err := repo.WriteFile("README.md", content); err != nil {
					logger.Error("Failed to write README.md: %v", err)
					os.Exit(1)
				}
				updatedFiles = append(updatedFiles, "README.md")
			}
		}
	}

	// Commit and push changes
	logger.Debug("Committing changes to repository")
	commitMsg := fmt.Sprintf("Update schedules: %s", strings.Join(updatedFiles, ", "))
	if err := repo.Commit(commitMsg); err != nil {
		logger.Error("Failed to commit changes: %v", err)
		os.Exit(1)
	}

	logger.Debug("Pushing changes to remote")
	if err := repo.Push(); err != nil {
		logger.Error("Failed to push changes: %v", err)
		os.Exit(1)
	}

	logger.Info("Successfully updated schedules: %s", strings.Join(updatedFiles, ", "))
	logger.Debug("DotCal application completed successfully")
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
