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

	// Calculate the start and end dates of the specified week
	weekStart := FirstDayOfISOWeek(year, week, m.timezone)
	weekEnd := weekStart.AddDate(0, 0, 7) // End of week (exclusive)

	logger.Debug("filtering events for week %d-%d (%s to %s)",
		year, week, weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02"))

	// Filter events to only include those within the specified week
	weekEvents := make([]Event, 0)
	for _, event := range events {
		// Skip weekend events
		if event.Start.Weekday() == time.Saturday || event.Start.Weekday() == time.Sunday {
			continue
		}

		// Check if event falls within the week
		if event.Start.Before(weekEnd) && event.End.After(weekStart) {
			weekEvents = append(weekEvents, event)
		}
	}

	logger.Debug("filtered %d events down to %d events for week %d",
		len(events), len(weekEvents), week)

	// Sort filtered events by start time
	sort.Slice(weekEvents, func(i, j int) bool {
		return weekEvents[i].Start.Before(weekEvents[j].Start)
	})

	// Merge events into slots
	for _, event := range weekEvents {
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

// FirstDayOfISOWeek returns the date of the first day (Monday) of the given ISO week
func FirstDayOfISOWeek(year int, week int, loc *time.Location) time.Time {
	// Start with January 4th which is always in week 1 of the ISO week year
	jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, loc)

	// Get the Monday of week 1
	_, w := jan4.ISOWeek()
	mon1 := jan4.AddDate(0, 0, -int(jan4.Weekday())+int(time.Monday))
	if w > 1 {
		mon1 = mon1.AddDate(0, 0, -7)
	}

	// Add weeks to get to our target week
	return mon1.AddDate(0, 0, (week-1)*7)
}

// mergeEventIntoSchedule merges a single event into the schedule
func (m *Merger) mergeEventIntoSchedule(schedule *WeekSchedule, event Event) {
	logger.Debug("processing event: ", event, " with status: ", event.Status)

	// Skip events outside business hours or on weekends
	if event.Start.Weekday() == time.Saturday || event.Start.Weekday() == time.Sunday {
		logger.Debug("skipping weekend event")
		return
	}

	// Calculate the start and end dates of the specified week
	// ISO week starts on Monday and ends on Sunday
	year, week := schedule.Year, schedule.Week
	weekStart := FirstDayOfISOWeek(year, week, m.timezone)
	weekEnd := weekStart.AddDate(0, 0, 7) // End of week (exclusive)

	// Skip events outside the specified week
	if event.Start.Before(weekStart) || event.Start.After(weekEnd) || event.Start.Equal(weekEnd) {
		logger.Debug("skipping event outside specified week")
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
