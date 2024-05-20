package channels

// This package defines the messages sent between our various channels, plus some utility functions for them.

import "fmt"

// A bundle of channels, to be passed around inside the controller.
type ChannelBundle struct {
	// Control messages.
	Control chan *ControlMessage

	// UPS info, to be consumed by VariableChangeMultiplexer
	Ups chan *UPSInfo

	// Variable updates to be consumed by metrics and translated to mqtt speak.
	Metrics       chan *UPSVariableUpdate
	MqttConverter chan *UPSVariableUpdate

	// MQTT updates to be consumed by the mqtt client
	Mqtt chan *MQTTUpdate
}

// A control message for the overall process.
type ControlMessage struct {
	Operation string
	Comment   string
}

func (cm ControlMessage) String() string {
	return fmt.Sprintf("Operation: %s, Comment: %s", cm.Operation, cm.Comment)
}

// A struct to describe an update to a UPS variable.
// to be consumed by the various goroutines via their channels.
type UPSVariableUpdate struct {
	Host    string
	UpsName string
	// the var name, in nut.dotted.format
	VarName string
	// the new value
	Content string
	// the previous value, if we have it.
	OldContent string
}

// MQTT

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
