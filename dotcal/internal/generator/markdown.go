package generator

import (
	"fmt"
	"strings"
	"time"

	"github.com/zach/dotcal/internal/calendar"
)

// Generator handles markdown schedule generation
type Generator struct {
	anonymizer interface {
		AnonymizeTitle(string) string
		AnonymizeLocation(string) string
	}
}

// NewGenerator creates a new markdown generator
func NewGenerator(anonymizer interface {
	AnonymizeTitle(string) string
	AnonymizeLocation(string) string
}) *Generator {
	return &Generator{
		anonymizer: anonymizer,
	}
}

// GenerateWeekSchedule creates a markdown schedule for a week
func (g *Generator) GenerateWeekSchedule(schedule *calendar.WeekSchedule) string {
	var sb strings.Builder

	// Get the dates for the week
	firstDay := g.getFirstDayOfWeek(schedule.Year, schedule.Week)
	lastDay := firstDay.AddDate(0, 0, 4) // Friday

	// Header
	sb.WriteString("# 📅 Weekly Availability Calendar\n\n")
	sb.WriteString("<div align=\"center\">\n\n")

	// Navigation
	prevWeek := firstDay.AddDate(0, 0, -7)
	nextWeek := firstDay.AddDate(0, 0, 7)
	sb.WriteString(fmt.Sprintf("[← Previous Week](./%d-W%02d.md) | ", prevWeek.Year(), g.getISOWeek(prevWeek)))
	sb.WriteString(fmt.Sprintf("Week of %s - %s, %d (Week %d)",
		firstDay.Format("January 2"),
		lastDay.Format("January 2"),
		firstDay.Year(),
		schedule.Week))
	sb.WriteString(fmt.Sprintf(" | [Next Week →](./%d-W%02d.md)\n\n", nextWeek.Year(), g.getISOWeek(nextWeek)))

	sb.WriteString("[Jump to Current Week](/README.md) | [View All Weeks](./calendar-index.md)\n")
	sb.WriteString("</div>\n\n")

	// Legend
	sb.WriteString("> 🟢 Available | 🟡 Tentative | 🔴 Busy \n\n")

	// Table header
	sb.WriteString("| Time | Monday | Tuesday | Wednesday | Thursday | Friday |\n")
	sb.WriteString("|:----:|:------:|:--------:|:---------:|:--------:|:------:|\n")

	// Time slots
	daySlots := schedule.Days[time.Monday] // Use Monday's slots as reference
	for i := range daySlots {
		slot := daySlots[i]
		timeStr := fmt.Sprintf("%s - %s",
			slot.Start.Format("3:04 PM"),
			slot.End.Format("3:04 PM"))

		sb.WriteString(fmt.Sprintf("| %s |", timeStr))

		// Add slots for each day
		for day := time.Monday; day <= time.Friday; day++ {
			daySlot := schedule.Days[day][i]
			sb.WriteString(" ")
			sb.WriteString(g.formatSlot(daySlot))
			sb.WriteString(" |")
		}
		sb.WriteString("\n")
	}

	// Footer
	sb.WriteString("\n---\n")
	sb.WriteString("### 📝 Legend\n")
	sb.WriteString(fmt.Sprintf("- All times are in %s (%s)\n",
		schedule.TimeZone.String(),
		g.formatTimezoneOffset(schedule.TimeZone)))
	sb.WriteString("- 🟢 Available: Click to schedule a meeting\n")
	sb.WriteString("- 🔴 Busy: Scheduled meeting or event\n")
	sb.WriteString("- 🟡 Tentative: Possibly available\n\n")

	sb.WriteString("### 🗓️ Quick Links\n")
	sb.WriteString("- [Add to Calendar](./calendar.ics)\n")
	sb.WriteString(fmt.Sprintf("- [View Month Overview](./%s.md)\n", firstDay.Format("2006-01")))
	sb.WriteString("- [Booking Guidelines](./booking-guidelines.md)\n\n")

	// Last updated
	sb.WriteString(fmt.Sprintf("### 🔄 Last Updated: %s\n",
		time.Now().In(schedule.TimeZone).Format("2006-01-02 15:04 MST")))

	return sb.String()
}

func (g *Generator) formatSlot(slot calendar.TimeSlot) string {
	var status, link string

	switch slot.Status {
	case calendar.StatusAvailable:
		status = "🟢"
		link = "https://cal.com"
	case calendar.StatusBusy:
		status = "🔴"
		if slot.Original != nil {
			link = g.anonymizer.AnonymizeLocation(slot.Original.Location)
		}
	case calendar.StatusTentative:
		status = "🟡"
		if slot.Original != nil {
			link = g.anonymizer.AnonymizeLocation(slot.Original.Location)
		}
	}

	title := "Book"
	if slot.Original != nil {
		title = g.anonymizer.AnonymizeTitle(slot.Original.Title)
	}

	if link != "" {
		return fmt.Sprintf("%s [%s](%s)", status, title, link)
	}
	return fmt.Sprintf("%s %s", status, title)
}

func (g *Generator) getFirstDayOfWeek(year, week int) time.Time {
	// Find the first day of the year
	jan1 := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)

	// Find the first Monday of the year
	firstMonday := jan1
	for firstMonday.Weekday() != time.Monday {
		firstMonday = firstMonday.AddDate(0, 0, 1)
	}

	// Add weeks to get to the desired week
	return firstMonday.AddDate(0, 0, (week-1)*7)
}

func (g *Generator) getISOWeek(t time.Time) int {
	_, week := t.ISOWeek()
	return week
}

func (g *Generator) formatTimezoneOffset(tz *time.Location) string {
	now := time.Now().In(tz)
	_, offset := now.Zone()

	hours := offset / 3600
	return fmt.Sprintf("UTC%+d", hours)
}
