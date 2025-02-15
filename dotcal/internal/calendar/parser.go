package calendar

import (
	"strings"
	"time"
)

// Parser handles parsing ICS calendar data
type Parser struct {
	timezone *time.Location
}

// NewParser creates a new calendar parser
func NewParser(timezone *time.Location) *Parser {
	if timezone == nil {
		timezone = time.UTC
	}
	return &Parser{timezone: timezone}
}

// Parse parses raw ICS data into events
func (p *Parser) Parse(data []byte) ([]Event, error) {
	lines := strings.Split(string(data), "\n")
	var events []Event
	var currentEvent *Event

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Handle line continuations
		for i+1 < len(lines) && strings.HasPrefix(lines[i+1], " ") {
			line += strings.TrimSpace(lines[i+1])
			i++
		}

		switch {
		case line == "BEGIN:VEVENT":
			currentEvent = &Event{}
		case line == "END:VEVENT":
			if currentEvent != nil {
				// Check SUMMARY for busy indication before adding event
				// Because some ics feeds might be private and publish
				// availability in summary field instead of title
				if strings.Contains(strings.ToUpper(currentEvent.Title), "BUSY") {
					currentEvent.Status = StatusBusy
				}
				// logger.Debug("Event: ", currentEvent)
				events = append(events, *currentEvent)
				currentEvent = nil
			}
		case strings.HasPrefix(line, "DTSTART"):
			if currentEvent != nil {
				currentEvent.Start = p.parseDateTime(line)
			}
		case strings.HasPrefix(line, "DTEND"):
			if currentEvent != nil {
				currentEvent.End = p.parseDateTime(line)
			}
		case strings.HasPrefix(line, "SUMMARY"):
			if currentEvent != nil {
				currentEvent.Title = p.parseText(line)
			}
		case strings.HasPrefix(line, "DESCRIPTION"):
			if currentEvent != nil {
				currentEvent.Description = p.parseText(line)
			}
		case strings.HasPrefix(line, "LOCATION"):
			if currentEvent != nil {
				currentEvent.Location = p.parseText(line)
			}
		case strings.HasPrefix(line, "STATUS"):
			// Outlook feed all entries are STATUS:CONFIRMED
			if currentEvent != nil {
				currentEvent.Status = p.parseStatus(line)
			}
		}
	}

	return events, nil
}

func (p *Parser) parseDateTime(line string) time.Time {
	parts := strings.Split(line, ":")
	if len(parts) != 2 {
		return time.Time{}
	}

	// Handle different datetime formats
	dt := parts[1]
	formats := []string{
		"20060102T150405Z", // UTC
		"20060102T150405",  // Local
		"YYYYMMDD",         // Date only
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, dt, p.timezone); err == nil {
			return t
		}
	}

	return time.Time{}
}

func (p *Parser) parseText(line string) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func (p *Parser) parseStatus(line string) Status {
	status := strings.ToUpper(p.parseText(line))
	// logger.Debug("status", status, line)
	switch status {
	case "TENTATIVE":
		return StatusTentative
	case "BUSY":
		return StatusBusy
	case "CONFIRMED":
		return StatusAvailable
	default:
		return StatusAvailable
	}
}
