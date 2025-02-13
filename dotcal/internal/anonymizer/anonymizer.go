package anonymizer

import (
	"strings"
)

// EventType represents the type of event for anonymization
type EventType string

const (
	TypeMeeting EventType = "Meeting"
)

// Anonymizer handles event anonymization
type Anonymizer struct {
	// Commented out for future use if needed
	// keywords map[string]EventType
}

// NewAnonymizer creates a new event anonymizer
func NewAnonymizer() *Anonymizer {
	return &Anonymizer{
		// Commented out for future use if needed
		/*
			keywords: map[string]EventType{
				"standup":       TypeMeeting,
				"stand-up":      TypeMeeting,
				"planning":      TypeMeeting,
				"plan":          TypeMeeting,
				"review":        TypeMeeting,
				"training":      TypeMeeting,
				"learn":         TypeMeeting,
				"sync":          TypeMeeting,
				"workshop":      TypeMeeting,
				"client":        TypeMeeting,
				"customer":      TypeMeeting,
				"retro":         TypeMeeting,
				"retrospective": TypeMeeting,
			},
		*/
	}
}

// AnonymizeTitle converts an event title to an anonymized version
func (a *Anonymizer) AnonymizeTitle(title string) string {
	// Commented out for future use if needed
	/*
		lowTitle := strings.ToLower(title)
		for keyword, eventType := range a.keywords {
			if strings.Contains(lowTitle, keyword) {
				return string(eventType)
			}
		}
	*/
	return string(TypeMeeting) // Always return "Meeting" for maximum privacy
}

// AnonymizeLocation returns a generic meeting link
func (a *Anonymizer) AnonymizeLocation(location string) string {
	if strings.HasPrefix(location, "http") {
		return "https://meet.xyz"
	}
	return ""
}
