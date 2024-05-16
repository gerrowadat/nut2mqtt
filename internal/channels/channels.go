package channels

// This package defines the messages sent between our various channels, plus some utility functions for them.

import "fmt"

// A control message for the overall process.
type ControlMessage struct {
	operation string
	comment   string
}

func (cm ControlMessage) String() string {
	return fmt.Sprintf("Operation: %s, Comment: %s", cm.operation, cm.comment)
}

func NewControlMessage(operation string, comment string) *ControlMessage {
	return &ControlMessage{operation: operation, comment: comment}
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
