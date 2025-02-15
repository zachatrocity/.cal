package calendar

import (
	"testing"
	"time"
)

func TestNewParser(t *testing.T) {
	t.Run("with nil timezone defaults to UTC", func(t *testing.T) {
		parser := NewParser(nil)
		if parser.timezone != time.UTC {
			t.Errorf("Expected UTC timezone, got %v", parser.timezone)
		}
	})

	t.Run("with specific timezone", func(t *testing.T) {
		nyc, _ := time.LoadLocation("America/New_York")
		parser := NewParser(nyc)
		if parser.timezone != nyc {
			t.Errorf("Expected NYC timezone, got %v", parser.timezone)
		}
	})
}

func TestParse(t *testing.T) {
	parser := NewParser(time.UTC)

	t.Run("basic event parsing", func(t *testing.T) {
		input := `BEGIN:VCALENDAR
BEGIN:VEVENT
DTSTART:20250215T100000Z
DTEND:20250215T110000Z
SUMMARY:Team Meeting
DESCRIPTION:Weekly sync
LOCATION:Conference Room
STATUS:CONFIRMED
END:VEVENT
END:VCALENDAR`

		events, err := parser.Parse([]byte(input))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("Expected 1 event, got %d", len(events))
		}

		event := events[0]
		expectedStart := time.Date(2025, 2, 15, 10, 0, 0, 0, time.UTC)
		if !event.Start.Equal(expectedStart) {
			t.Errorf("Expected start time %v, got %v", expectedStart, event.Start)
		}

		expectedEnd := time.Date(2025, 2, 15, 11, 0, 0, 0, time.UTC)
		if !event.End.Equal(expectedEnd) {
			t.Errorf("Expected end time %v, got %v", expectedEnd, event.End)
		}

		if event.Title != "Team Meeting" {
			t.Errorf("Expected title 'Team Meeting', got '%s'", event.Title)
		}

		if event.Description != "Weekly sync" {
			t.Errorf("Expected description 'Weekly sync', got '%s'", event.Description)
		}

		if event.Location != "Conference Room" {
			t.Errorf("Expected location 'Conference Room', got '%s'", event.Location)
		}

		if event.Status != StatusAvailable {
			t.Errorf("Expected status Available, got %v", event.Status)
		}
	})

	t.Run("line continuation", func(t *testing.T) {
		input := `BEGIN:VCALENDAR
BEGIN:VEVENT
DTSTART:20250215T100000Z
DESCRIPTION:This is a very long description that spans
 multiple lines in the ICS file but should be treated
 as a single continuous string
END:VEVENT
END:VCALENDAR`

		events, err := parser.Parse([]byte(input))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("Expected 1 event, got %d", len(events))
		}

		expected := "This is a very long description that spans multiple lines in the ICS file but should be treated as a single continuous string"
		if events[0].Description != expected {
			t.Errorf("Expected description '%s', got '%s'", expected, events[0].Description)
		}
	})

	t.Run("multiple events", func(t *testing.T) {
		input := `BEGIN:VCALENDAR
BEGIN:VEVENT
DTSTART:20250215T100000Z
DTEND:20250215T110000Z
SUMMARY:Event 1
END:VEVENT
BEGIN:VEVENT
DTSTART:20250215T120000Z
DTEND:20250215T130000Z
SUMMARY:Event 2
END:VEVENT
END:VCALENDAR`

		events, err := parser.Parse([]byte(input))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(events) != 2 {
			t.Fatalf("Expected 2 events, got %d", len(events))
		}

		if events[0].Title != "Event 1" {
			t.Errorf("Expected first event title 'Event 1', got '%s'", events[0].Title)
		}
		if events[1].Title != "Event 2" {
			t.Errorf("Expected second event title 'Event 2', got '%s'", events[1].Title)
		}
	})

	t.Run("different status values", func(t *testing.T) {
		tests := []struct {
			status   string
			expected Status
		}{
			{"TENTATIVE", StatusTentative},
			{"BUSY", StatusBusy},
			{"CONFIRMED", StatusAvailable},
			{"FREE", StatusAvailable},
			{"", StatusAvailable},
		}

		for _, tc := range tests {
			input := `BEGIN:VCALENDAR
BEGIN:VEVENT
DTSTART:20250215T100000Z
STATUS:` + tc.status + `
END:VEVENT
END:VCALENDAR`

			events, err := parser.Parse([]byte(input))
			if err != nil {
				t.Fatalf("Unexpected error for status %s: %v", tc.status, err)
			}
			if len(events) != 1 {
				t.Fatalf("Expected 1 event for status %s, got %d", tc.status, len(events))
			}
			if events[0].Status != tc.expected {
				t.Errorf("For status '%s': expected %v, got %v", tc.status, tc.expected, events[0].Status)
			}
		}
	})

	t.Run("busy in summary", func(t *testing.T) {
		input := `BEGIN:VCALENDAR
BEGIN:VEVENT
DTSTART:20250215T100000Z
SUMMARY:BUSY: Important Meeting
STATUS:CONFIRMED
END:VEVENT
END:VCALENDAR`

		events, err := parser.Parse([]byte(input))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("Expected 1 event, got %d", len(events))
		}
		if events[0].Status != StatusBusy {
			t.Errorf("Expected status Busy due to summary, got %v", events[0].Status)
		}
	})

	t.Run("different datetime formats", func(t *testing.T) {
		tests := []struct {
			dtstart  string
			expected time.Time
		}{
			{"20250215T100000Z", time.Date(2025, 2, 15, 10, 0, 0, 0, time.UTC)},
			{"20250215T100000", time.Date(2025, 2, 15, 10, 0, 0, 0, time.UTC)},
		}

		for _, tc := range tests {
			input := `BEGIN:VCALENDAR
BEGIN:VEVENT
DTSTART:` + tc.dtstart + `
END:VEVENT
END:VCALENDAR`

			events, err := parser.Parse([]byte(input))
			if err != nil {
				t.Fatalf("Unexpected error for datetime %s: %v", tc.dtstart, err)
			}
			if len(events) != 1 {
				t.Fatalf("Expected 1 event for datetime %s, got %d", tc.dtstart, len(events))
			}
			if !events[0].Start.Equal(tc.expected) {
				t.Errorf("For datetime '%s': expected %v, got %v", tc.dtstart, tc.expected, events[0].Start)
			}
		}
	})
}
