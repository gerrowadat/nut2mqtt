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
	// The full topic - note, --mqtt-topic-base is prepended to this.
	Topic string
	// The update content
	Content string
	// The previous version of the content, if we have it.
	OldContent string
}

// UPS Info
type UPSInfo struct {
	// Name of the UPS as configured in nut
	Name string
	// hostname of the connected machine
	Host string
	// Description, as returned by nut
	Description string
	// All variable returned by LIST VAR
	Vars map[string]string
}
