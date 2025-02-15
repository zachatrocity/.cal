package calendar

import (
	"testing"
	"time"
)

func TestNewMerger(t *testing.T) {
	t.Run("with nil timezone defaults to UTC", func(t *testing.T) {
		merger := NewMerger(nil)
		if merger.timezone != time.UTC {
			t.Errorf("Expected UTC timezone, got %v", merger.timezone)
		}
	})

	t.Run("with specific timezone", func(t *testing.T) {
		nyc, _ := time.LoadLocation("America/New_York")
		merger := NewMerger(nyc)
		if merger.timezone != nyc {
			t.Errorf("Expected NYC timezone, got %v", merger.timezone)
		}
	})
}

func TestMergeEvents(t *testing.T) {
	// Use a fixed date for consistent testing
	baseDate := time.Date(2025, 2, 17, 0, 0, 0, 0, time.UTC) // A Monday
	merger := NewMerger(time.UTC)

	t.Run("empty events list", func(t *testing.T) {
		schedule := merger.MergeEvents(nil, 2025, 8)
		if len(schedule.Days) != 5 {
			t.Errorf("Expected 5 days, got %d", len(schedule.Days))
		}
		// Verify each day has 16 slots (8 hours * 2 slots per hour)
		for day := time.Monday; day <= time.Friday; day++ {
			slots := schedule.Days[day]
			if len(slots) != 16 {
				t.Errorf("Expected 16 slots for %v, got %d", day, len(slots))
			}
			// Verify all slots are available
			for _, slot := range slots {
				if slot.Status != StatusAvailable {
					t.Errorf("Expected available status, got %v", slot.Status)
				}
			}
		}
	})

	t.Run("single event during business hours", func(t *testing.T) {
		events := []Event{
			{
				Start:  baseDate.Add(10 * time.Hour), // 10 AM
				End:    baseDate.Add(11 * time.Hour), // 11 AM
				Status: StatusBusy,
			},
		}
		schedule := merger.MergeEvents(events, 2025, 8)

		// Check affected slots
		mondaySlots := schedule.Days[time.Monday]
		// 10:00-10:30 slot (index 2) and 10:30-11:00 slot (index 3) should be busy
		for i := 2; i <= 3; i++ {
			if mondaySlots[i].Status != StatusBusy {
				t.Errorf("Expected busy status for slot %d, got %v", i, mondaySlots[i].Status)
			}
		}
	})

	t.Run("weekend events are skipped", func(t *testing.T) {
		saturdayEvent := Event{
			Start:  baseDate.Add(144 * time.Hour), // Next Saturday
			End:    baseDate.Add(145 * time.Hour),
			Status: StatusBusy,
		}
		schedule := merger.MergeEvents([]Event{saturdayEvent}, 2025, 8)

		// Verify all slots remain available
		for day := time.Monday; day <= time.Friday; day++ {
			for _, slot := range schedule.Days[day] {
				if slot.Status != StatusAvailable {
					t.Errorf("Weekend event affected weekday slot: %v", slot)
				}
			}
		}
	})

	t.Run("overlapping events with different statuses", func(t *testing.T) {
		events := []Event{
			{
				Start:  baseDate.Add(10 * time.Hour),                // 10 AM
				End:    baseDate.Add(10*time.Hour + 30*time.Minute), // 10:30 AM
				Status: StatusTentative,
			},
			{
				Start:  baseDate.Add(10 * time.Hour),                // 10 AM
				End:    baseDate.Add(10*time.Hour + 30*time.Minute), // 10:30 AM
				Status: StatusBusy,
			},
		}
		schedule := merger.MergeEvents(events, 2025, 8)

		// Check the 10:00-10:30 slot (index 2)
		slot := schedule.Days[time.Monday][2]
		if slot.Status != StatusBusy {
			t.Errorf("Expected busy status due to priority, got %v", slot.Status)
		}
	})

	t.Run("event spanning multiple slots", func(t *testing.T) {
		events := []Event{
			{
				Start:  baseDate.Add(10 * time.Hour), // 10 AM
				End:    baseDate.Add(12 * time.Hour), // 12 PM
				Status: StatusBusy,
			},
		}
		schedule := merger.MergeEvents(events, 2025, 8)

		// Check all affected slots (10:00-12:00, 4 slots)
		mondaySlots := schedule.Days[time.Monday]
		for i := 2; i <= 5; i++ {
			if mondaySlots[i].Status != StatusBusy {
				t.Errorf("Expected busy status for slot %d, got %v", i, mondaySlots[i].Status)
			}
		}
	})

	t.Run("status priority handling", func(t *testing.T) {
		events := []Event{
			{
				Start:  baseDate.Add(10 * time.Hour),
				End:    baseDate.Add(11 * time.Hour),
				Status: StatusAvailable,
			},
			{
				Start:  baseDate.Add(10 * time.Hour),
				End:    baseDate.Add(11 * time.Hour),
				Status: StatusTentative,
			},
			{
				Start:  baseDate.Add(10 * time.Hour),
				End:    baseDate.Add(11 * time.Hour),
				Status: StatusBusy,
			},
		}
		schedule := merger.MergeEvents(events, 2025, 8)

		// Check affected slots (10:00-11:00, 2 slots)
		mondaySlots := schedule.Days[time.Monday]
		for i := 2; i <= 3; i++ {
			if mondaySlots[i].Status != StatusBusy {
				t.Errorf("Expected busy status for slot %d due to priority, got %v", i, mondaySlots[i].Status)
			}
			if mondaySlots[i].Original == nil {
				t.Error("Expected Original event reference to be set")
			}
		}
	})
}
