package calendar

import (
	"sort"
	"time"

	"github.com/zach/dotcal/internal/logger"
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

	// Use a reference date for the slots (actual year/month/day will be set during merge)
	start := time.Date(2000, 1, 1, 9, 0, 0, 0, m.timezone)

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
	logger.Debug("processing event: ", event, " with status: ", event.Status)

	// Skip events outside business hours or on weekends
	if event.Start.Weekday() == time.Saturday || event.Start.Weekday() == time.Sunday {
		logger.Debug("skipping weekend event")
		return
	}

	daySlots := schedule.Days[event.Start.Weekday()]
	eventDate := time.Date(event.Start.Year(), event.Start.Month(), event.Start.Day(), 0, 0, 0, 0, m.timezone)

	for i := range daySlots {
		slot := &daySlots[i]

		// Adjust slot times to match event date
		slotStart := time.Date(
			eventDate.Year(), eventDate.Month(), eventDate.Day(),
			slot.Start.Hour(), slot.Start.Minute(), 0, 0,
			m.timezone,
		)
		slotEnd := slotStart.Add(30 * time.Minute)

		// logger.Debug("event: ", event, " status: ", event.Status)
		// logger.Debug("slot before: ", slot)
		// logger.Debug("adjusted slot time - start: ", slotStart, " end: ", slotEnd)

		// Check if event overlaps with this slot using adjusted times
		if event.Start.Before(slotEnd) && event.End.After(slotStart) {
			logger.Debug("overlap detected with status: ", event.Status)
			// Update slot based on status priority
			switch {
			case event.Status == StatusBusy:
				// Busy always takes precedence
				slot.Status = StatusBusy
				slot.Original = &event
			case event.Status == StatusTentative && slot.Status != StatusBusy:
				// Tentative takes precedence over Available
				slot.Status = StatusTentative
				slot.Original = &event
			case event.Status == StatusAvailable:
				// For Available events, always update the Original reference
				// but keep the slot's current status
				slot.Original = &event
				slot.Status = StatusAvailable
			}
		}
	}
}
