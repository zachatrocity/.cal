# dotcal Project Context

## Core Concepts

dotcal is a self-hosted calendar aggregation service that:
- Combines multiple ICS feeds into a unified schedule
- Anonymizes all events (shown only as "Meeting" for privacy)
- Publishes schedules to GitHub as markdown files
- Updates on a configurable schedule

## Architecture

### Directory Structure
```
dotcal/
├── dotcal/                  # Go application
│   ├── cmd/server/         # Main application entry point
│   └── internal/           # Internal packages
│       ├── dotcal/        # Calendar processing
│       ├── generator/     # Markdown generation
│       └── git/          # Git operations
├── docker-compose.yml      # Service orchestration
└── template.md            # Schedule template
```

### Components

1. Calendar Processing (internal/dotcal):
   - Fetches ICS feeds (URLs or files)
   - Parses calendar data
   - Merges multiple calendars
   - Handles timezone conversions
   - 30-minute time slot granularity

2. Schedule Generation (internal/generator):
   - Markdown-based output
   - Weekly schedule view (Mon-Fri)
   - Time slots from 9 AM to 5 PM
   - Status indicators (🟢🔴🟡)

3. Git Operations (internal/git):
   - Repository cloning and updates
   - File management
   - Automated commits and pushes

### Schedule Organization

- README.md: Current week's schedule
- future/YYYY-WXX.md: Upcoming weeks
- past/YYYY-WXX.md: Past weeks

Friday evening automation:
- Moves current README.md to past/
- Updates README.md with next week's schedule

### Configuration

Required Environment Variables:
- GITHUB_REPO: Repository URL (git@github.com:username/repo.git)
- ICS_FEEDS: Comma-separated feed URLs/paths

Optional Environment Variables:
- GITHUB_BRANCH: Branch name (default: main)
- TIMEZONE: Schedule timezone (default: UTC)
- SYNC_SCHEDULE: Update frequency (default: */30 * * * *)
- SCHEDULE_MONTHS: Number of months to generate (default: 3)
- SSH_KEY_FILE: SSH key path (default: ~/.ssh/id_rsa)

## Design Decisions

1. Privacy First:
   - All events anonymized as "Busy"
   - No event details exposed

2. Static Output:
   - Markdown files for simplicity
   - GitHub as the hosting platform
   - No dynamic server required

3. Schedule Range:
   - Keeps 1 month of history
   - Configurable future months
   - Weekly granularity

4. Docker-based Deployment:
   - Self-contained service
   - Simple environment configuration
   - Host-mapped SSH key and repo data

5. Test Driven Development:
   - Where possible create unit tests
   - Create mock data where necessary
   - Ensure go coverage is measured

## Implementation Guidelines

When implementing new features:
1. Maintain strict privacy (no event details)
2. Keep markdown output simple and clean
3. Use Go's standard library when possible
4. Follow existing package structure
5. Add clear error messages for configuration issues
6. Consider timezone handling in all date/time operations

## Testing Considerations

1. Calendar Processing:
   - Various ICS formats
   - Timezone conversions
   - Event overlaps

2. Schedule Generation:
   - Week transitions
   - Friday evening updates
   - Archive organization

3. Error Handling:
   - Missing required variables
   - Invalid feed URLs
   - Git operation failures
