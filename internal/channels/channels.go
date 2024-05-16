package channels

// This package defines the messages sent between our various channels, plus some utility functions for them.

import "fmt"

// A control message for the overall process.
type ControlMessage struct {
	Operation string
	Comment   string
}

func (cm ControlMessage) String() string {
	return fmt.Sprintf("Operation: %s, Comment: %s", cm.Operation, cm.Comment)
}

// MQTT

// A struct to describe an MQTT update to be made.
type MQTTUpdate struct {
	Topic      string
	Content    string
	OldContent string
}

// UPS Info
type UPSInfo struct {
	Name        string
	Description string
	Vars        map[string]string
}
