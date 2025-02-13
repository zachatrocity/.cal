package calendar

import (
	"time"
)

// Status represents the availability status of a time slot
type Status string

const (
	StatusAvailable Status = "available"
	StatusBusy      Status = "busy"
	StatusTentative Status = "tentative"
)

// Event represents a calendar event
type Event struct {
	Start       time.Time
	End         time.Time
	Status      Status
	Title       string
	Description string
	Location    string
}

// TimeSlot represents a 30-minute slot in the schedule
type TimeSlot struct {
	Start    time.Time
	End      time.Time
	Status   Status
	Original *Event // Reference to original event if not available
}

// Feed represents a calendar feed source
type Feed struct {
	ID       string
	Source   string // URL or file path
	IsURL    bool
	TimeZone *time.Location
}

// Schedule represents a processed calendar schedule
type Schedule struct {
	TimeZone *time.Location
	Slots    []TimeSlot
}

// WeekSchedule represents a full week of time slots
type WeekSchedule struct {
	Year     int
	Week     int
	TimeZone *time.Location
	Days     map[time.Weekday][]TimeSlot
}
