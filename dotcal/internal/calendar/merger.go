package calendar

import (
	"sort"
	"time"
)

// Merger handles merging multiple calendars into a unified schedule
type Merger struct {
	timezone *time.Location
}

// NewMerger creates a new calendar merger
func NewMerger(timezone *time.Location) *Merger {
	if timezone == nil {
		timezone = time.UTC
	}
	return &Merger{timezone: timezone}
}

// MergeEvents combines multiple event lists into a unified weekly schedule
func (m *Merger) MergeEvents(events []Event, year int, week int) *WeekSchedule {
	schedule := &WeekSchedule{
		Year:     year,
		Week:     week,
		TimeZone: m.timezone,
		Days:     make(map[time.Weekday][]TimeSlot),
	}

	// Initialize empty slots for each day
	for day := time.Monday; day <= time.Friday; day++ {
		schedule.Days[day] = m.createDaySlots()
	}

	// Sort events by start time
	sort.Slice(events, func(i, j int) bool {
		return events[i].Start.Before(events[j].Start)
	})

	// Merge events into slots
	for _, event := range events {
		m.mergeEventIntoSchedule(schedule, event)
	}

	return schedule
}

// createDaySlots creates empty 30-minute slots for a day (9 AM to 5 PM)
func (m *Merger) createDaySlots() []TimeSlot {
	slots := make([]TimeSlot, 0, 16) // 8 hours * 2 slots per hour

	// Start at 9 AM
	start := time.Date(0, 0, 0, 9, 0, 0, 0, m.timezone)

	for i := 0; i < 16; i++ {
		slotStart := start.Add(time.Duration(i) * 30 * time.Minute)
		slots = append(slots, TimeSlot{
			Start:  slotStart,
			End:    slotStart.Add(30 * time.Minute),
			Status: StatusAvailable,
		})
	}

	return slots
}

// mergeEventIntoSchedule merges a single event into the schedule
func (m *Merger) mergeEventIntoSchedule(schedule *WeekSchedule, event Event) {
	// Skip events outside business hours or on weekends
	if event.Start.Weekday() == time.Saturday || event.Start.Weekday() == time.Sunday {
		return
	}

	daySlots := schedule.Days[event.Start.Weekday()]
	for i := range daySlots {
		slot := &daySlots[i]

		// Check if event overlaps with this slot
		if event.Start.Before(slot.End) && event.End.After(slot.Start) {
			// Update slot status based on event priority
			if event.Status == StatusBusy ||
				(event.Status == StatusTentative && slot.Status == StatusAvailable) {
				slot.Status = event.Status
				slot.Original = &event
			}
		}
	}
}
