package generator

import (
	"strings"
	"testing"
	"time"

	"github.com/zach/dotcal/internal/calendar"
)

func TestNewGenerator(t *testing.T) {
	tests := []struct {
		name        string
		templateDir string
		wantErr     bool
	}{
		{
			name:        "valid template directory",
			templateDir: "../templates",
			wantErr:     false,
		},
		{
			name:        "invalid template directory",
			templateDir: "nonexistent",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := NewGenerator(tt.templateDir)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if g == nil {
				t.Error("expected generator, got nil")
			}
		})
	}
}

func TestGenerateWeekSchedule(t *testing.T) {
	g, err := NewGenerator("../templates")
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	// Create a test schedule
	// Initialize slots for all weekdays
	baseSlots := []calendar.TimeSlot{
		{
			Start:  time.Date(2025, 2, 10, 9, 0, 0, 0, time.UTC),
			End:    time.Date(2025, 2, 10, 10, 0, 0, 0, time.UTC),
			Status: calendar.StatusAvailable,
		},
		{
			Start:  time.Date(2025, 2, 10, 10, 0, 0, 0, time.UTC),
			End:    time.Date(2025, 2, 10, 11, 0, 0, 0, time.UTC),
			Status: calendar.StatusBusy,
		},
	}

	schedule := &calendar.WeekSchedule{
		Year:     2025,
		Week:     7,
		TimeZone: time.UTC,
		Days:     make(map[time.Weekday][]calendar.TimeSlot),
	}

	// Copy slots to each weekday, adjusting dates
	for day := time.Monday; day <= time.Friday; day++ {
		daySlots := make([]calendar.TimeSlot, len(baseSlots))
		for i, slot := range baseSlots {
			daySlot := slot
			dayOffset := int(day - time.Monday)
			daySlot.Start = slot.Start.AddDate(0, 0, dayOffset)
			daySlot.End = slot.End.AddDate(0, 0, dayOffset)
			daySlots[i] = daySlot
		}
		schedule.Days[day] = daySlots
	}

	// Override some slots with different statuses for variety
	schedule.Days[time.Tuesday][0].Status = calendar.StatusTentative
	schedule.Days[time.Wednesday][1].Status = calendar.StatusTentative
	schedule.Days[time.Thursday][1].Status = calendar.StatusBusy
	schedule.Days[time.Friday][0].Status = calendar.StatusTentative

	output, err := g.GenerateWeekSchedule(schedule)
	if err != nil {
		t.Fatalf("failed to generate schedule: %v", err)
	}

	// Print output for debugging
	t.Logf("Generated output:\n%s", output)

	// Test output contains expected elements
	expectedElements := []string{
		"📅 Weekly Availability Calendar",
		"Week of February 10 - February 14",
		"9:00 AM - 10:00 AM",
		"🟢 Available",
		"🔴 Busy",
		"🟡 Tentative",
		"[← Previous Week]",
		"[Next Week →]",
		"[Jump to Current Week]",
		"[View All Weeks]",
		"All times are in UTC",
		"Last Updated:",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q", expected)
		}
	}
}

func TestBuildDaySlot(t *testing.T) {
	g := &Generator{}

	tests := []struct {
		name     string
		slot     calendar.TimeSlot
		expected DaySlotData
	}{
		{
			name: "available slot",
			slot: calendar.TimeSlot{Status: calendar.StatusAvailable},
			expected: DaySlotData{
				Status: "🟢",
				Title:  "Available",
				Link:   "https://cal.com",
			},
		},
		{
			name: "busy slot",
			slot: calendar.TimeSlot{Status: calendar.StatusBusy},
			expected: DaySlotData{
				Status: "🔴",
				Title:  "Busy",
				Link:   "",
			},
		},
		{
			name: "tentative slot",
			slot: calendar.TimeSlot{Status: calendar.StatusTentative},
			expected: DaySlotData{
				Status: "🟡",
				Title:  "Tentative",
				Link:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := g.buildDaySlot(tt.slot)
			if result.Status != tt.expected.Status ||
				result.Title != tt.expected.Title ||
				result.Link != tt.expected.Link {
				t.Errorf("buildDaySlot() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBuildNavigation(t *testing.T) {
	g := &Generator{}
	year := 2025
	week := 7

	nav := g.buildNavigation(year, week)

	if nav.PrevLink == "" {
		t.Error("expected non-empty PrevLink")
	}
	if nav.NextLink == "" {
		t.Error("expected non-empty NextLink")
	}
	if nav.CurrentLink != "/README.md" {
		t.Errorf("expected CurrentLink to be /README.md, got %s", nav.CurrentLink)
	}
	if nav.IndexLink != "/calendar-index.md" {
		t.Errorf("expected IndexLink to be /calendar-index.md, got %s", nav.IndexLink)
	}
}

func TestTemplateFuncs(t *testing.T) {
	g := &Generator{}
	funcs := g.templateFuncs()

	// Test formatTime
	formatTime := funcs["formatTime"].(func(time.Time) string)
	testTime := time.Date(2025, 2, 15, 14, 30, 0, 0, time.UTC)
	if got := formatTime(testTime); got != "2:30 PM" {
		t.Errorf("formatTime() = %v, want %v", got, "2:30 PM")
	}

	// Test formatDate
	formatDate := funcs["formatDate"].(func(time.Time) string)
	if got := formatDate(testTime); got != "February 15" {
		t.Errorf("formatDate() = %v, want %v", got, "February 15")
	}

	// Test formatStatus
	formatStatus := funcs["formatStatus"].(func(DaySlotData) string)
	tests := []struct {
		slot DaySlotData
		want string
	}{
		{
			slot: DaySlotData{Status: "🟢", Title: "Available", Link: "https://cal.com"},
			want: "🟢 [Available](https://cal.com)",
		},
		{
			slot: DaySlotData{Status: "🔴", Title: "Busy"},
			want: "🔴 Busy",
		},
	}

	for _, tt := range tests {
		if got := formatStatus(tt.slot); got != tt.want {
			t.Errorf("formatStatus() = %v, want %v", got, tt.want)
		}
	}

	// Test timezoneOffset
	timezoneOffset := funcs["timezoneOffset"].(func(*time.Location) string)
	if got := timezoneOffset(time.UTC); got != "UTC+0" {
		t.Errorf("timezoneOffset() = %v, want %v", got, "UTC+0")
	}
}

func TestFirstDayOfISOWeek(t *testing.T) {
	tests := []struct {
		year     int
		week     int
		wantDate time.Time
	}{
		{
			year:     2025,
			week:     1,
			wantDate: time.Date(2024, 12, 30, 0, 0, 0, 0, time.UTC),
		},
		{
			year:     2025,
			week:     7,
			wantDate: time.Date(2025, 2, 10, 0, 0, 0, 0, time.UTC),
		},
		{
			year:     2025,
			week:     9,
			wantDate: time.Date(2025, 2, 24, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		got := calendar.FirstDayOfISOWeek(tt.year, tt.week, time.UTC)
		if !got.Equal(tt.wantDate) {
			t.Errorf("FirstDayOfISOWeek(%d, %d, UTC) = %v, want %v",
				tt.year, tt.week, got, tt.wantDate)
		}
	}
}
