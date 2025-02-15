package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/zach/dotcal/internal/calendar"
)

// Generator handles markdown schedule generation
type Generator struct {
	templateDir string
	templates   map[string]*template.Template
}

// TemplateData holds common template data
type TemplateData struct {
	Navigation  NavigationData
	TimeZone    *time.Location
	LastUpdated string
}

// NavigationData holds navigation links
type NavigationData struct {
	PrevLink    string
	NextLink    string
	CurrentLink string
	IndexLink   string
}

// WeekTemplateData holds data for weekly view
type WeekTemplateData struct {
	TemplateData
	Schedule  *calendar.WeekSchedule
	TimeSlots []TimeSlotData
	StartDate time.Time
	EndDate   time.Time
}

// TimeSlotData represents a single time slot
type TimeSlotData struct {
	Time     string
	DaySlots []DaySlotData
}

// DaySlotData represents a slot for a specific day
type DaySlotData struct {
	Status string
	Title  string
	Link   string
}

// NewGenerator creates a new markdown generator
func NewGenerator(templateDir string) (*Generator, error) {
	g := &Generator{
		templateDir: templateDir,
		templates:   make(map[string]*template.Template),
	}

	if err := g.loadTemplates(); err != nil {
		return nil, fmt.Errorf("loading templates: %w", err)
	}

	return g, nil
}

// loadTemplates loads all template files
func (g *Generator) loadTemplates() error {
	// Only load weekly template for now
	name := "weekly"
	defaultPath := filepath.Join(g.templateDir, "default", name+".md.tmpl")
	customPath := filepath.Join(g.templateDir, "custom", name+".md.tmpl")

	var templatePath string
	if _, err := os.Stat(customPath); err == nil {
		templatePath = customPath
	} else {
		templatePath = defaultPath
	}

	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("reading template %s: %w", name, err)
	}

	tmpl, err := template.New(name).Funcs(g.templateFuncs()).Parse(string(content))
	if err != nil {
		return fmt.Errorf("parsing template %s: %w", name, err)
	}

	g.templates[name] = tmpl

	return nil
}

// GenerateWeekSchedule creates a markdown schedule for a week
func (g *Generator) GenerateWeekSchedule(schedule *calendar.WeekSchedule) (string, error) {
	startDate := g.getFirstDayOfWeek(schedule.Year, schedule.Week)
	endDate := startDate.AddDate(0, 0, 4) // Friday

	data := WeekTemplateData{
		StartDate: startDate,
		EndDate:   endDate,
		TemplateData: TemplateData{
			Navigation:  g.buildNavigation(schedule.Year, schedule.Week),
			TimeZone:    schedule.TimeZone,
			LastUpdated: time.Now().In(schedule.TimeZone).Format("2006-01-02 15:04 MST"),
		},
		Schedule:  schedule,
		TimeSlots: g.buildTimeSlots(schedule),
	}

	var output strings.Builder
	if err := g.templates["weekly"].Execute(&output, data); err != nil {
		return "", fmt.Errorf("executing weekly template: %w", err)
	}

	return output.String(), nil
}

// buildTimeSlots converts schedule slots into template data
func (g *Generator) buildTimeSlots(schedule *calendar.WeekSchedule) []TimeSlotData {
	var slots []TimeSlotData
	daySlots := schedule.Days[time.Monday] // Use Monday's slots as reference

	for i := range daySlots {
		slot := daySlots[i]
		timeStr := fmt.Sprintf("%s - %s",
			slot.Start.Format("3:04 PM"),
			slot.End.Format("3:04 PM"))

		var daySlots []DaySlotData
		for day := time.Monday; day <= time.Friday; day++ {
			daySlot := schedule.Days[day][i]
			daySlots = append(daySlots, g.buildDaySlot(daySlot))
		}

		slots = append(slots, TimeSlotData{
			Time:     timeStr,
			DaySlots: daySlots,
		})
	}

	return slots
}

// buildDaySlot converts a calendar time slot into template data
func (g *Generator) buildDaySlot(slot calendar.TimeSlot) DaySlotData {
	var status, title, link string

	switch slot.Status {
	case calendar.StatusAvailable:
		status = "🟢"
		title = "Available"
		link = "https://cal.com"
	case calendar.StatusBusy:
		status = "🔴"
		title = "Busy"
	case calendar.StatusTentative:
		status = "🟡"
		title = "Tentative"
	}

	return DaySlotData{
		Status: status,
		Title:  title,
		Link:   link,
	}
}

// buildNavigation creates navigation links for templates
func (g *Generator) buildNavigation(year, week int) NavigationData {
	prevWeek := g.getFirstDayOfWeek(year, week).AddDate(0, 0, -7)
	nextWeek := g.getFirstDayOfWeek(year, week).AddDate(0, 0, 7)
	now := time.Now()

	prevYear, prevWeekNum := prevWeek.ISOWeek()
	nextYear, nextWeekNum := nextWeek.ISOWeek()

	var prevPath, nextPath string
	if prevWeek.Before(now) {
		prevPath = fmt.Sprintf("/past/%d-W%02d.md", prevYear, prevWeekNum)
	} else {
		prevPath = fmt.Sprintf("/future/%d-W%02d.md", prevYear, prevWeekNum)
	}

	if nextWeek.Before(now) {
		nextPath = fmt.Sprintf("/past/%d-W%02d.md", nextYear, nextWeekNum)
	} else {
		nextPath = fmt.Sprintf("/future/%d-W%02d.md", nextYear, nextWeekNum)
	}

	return NavigationData{
		PrevLink:    prevPath,
		NextLink:    nextPath,
		CurrentLink: "/README.md",
		IndexLink:   "/calendar-index.md",
	}
}

// templateFuncs returns template helper functions
func (g *Generator) templateFuncs() template.FuncMap {
	return template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("3:04 PM")
		},
		"formatDate": func(t time.Time) string {
			return t.Format("January 2")
		},
		"formatStatus": func(slot DaySlotData) string {
			if slot.Link != "" {
				return fmt.Sprintf("%s [%s](%s)", slot.Status, slot.Title, slot.Link)
			}
			return fmt.Sprintf("%s %s", slot.Status, slot.Title)
		},
		"timezoneOffset": g.formatTimezoneOffset,
	}
}

// Helper methods from original implementation
func (g *Generator) getFirstDayOfWeek(year, week int) time.Time {
	// Find the first day of the year
	jan1 := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)

	// Get the offset to the first Monday of the year
	offset := int(time.Monday - jan1.Weekday())
	if offset > 0 {
		offset -= 7
	}

	// Get the first Monday of the year
	firstMonday := jan1.AddDate(0, 0, offset)

	// Add weeks to get to the target week
	return firstMonday.AddDate(0, 0, (week-1)*7)
}

func (g *Generator) formatTimezoneOffset(tz *time.Location) string {
	now := time.Now().In(tz)
	_, offset := now.Zone()
	hours := offset / 3600
	return fmt.Sprintf("UTC%+d", hours)
}
