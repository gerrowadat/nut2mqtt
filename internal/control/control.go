// Program Control nonsense.

package control

import (
	"fmt"
	"sync"

	channels "github.com/gerrowadat/nut2mqtt/internal/channels"
)

type Controller struct {
	control_chan chan *channels.ControlMessage
	mqtt_chan    chan *channels.MQTTUpdate
	mqtt_topic   string
}

func NewController(mqtt_change_chan chan *channels.MQTTUpdate, mqtt_topic string) Controller {
	return Controller{
		control_chan: make(chan *channels.ControlMessage),
		mqtt_chan:    mqtt_change_chan,
		mqtt_topic:   mqtt_topic}
}

func (c Controller) Startup(comment string, args ...interface{}) {
	comment = fmt.Sprintf("Startup: "+comment, args...)
	c.control_chan <- channels.NewControlMessage("startup", comment)
}

func (c Controller) Shutdown(comment string, args ...interface{}) {
	comment = fmt.Sprintf("Shutdown: "+comment, args...)
	c.control_chan <- channels.NewControlMessage("shutdown", comment)
}

func (c *Controller) ControlMessageConsumer(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		msg := <-c.control_chan
		fmt.Println("Processing Control message: ", msg.String())
		switch msg.Operation {
		case "startup":
			c.mqtt_chan <- &channels.MQTTUpdate{Topic: c.mqtt_topic + "/state", Content: "online"}
		case "shutdown":
			c.mqtt_chan <- &channels.MQTTUpdate{Topic: c.mqtt_topic + "/state", Content: "offline"}
			// returning will exit the consumer, and process will end.
			return
		default:
			fmt.Println("Unknown operation on control channel: ", msg.Operation)
		}
	}
}
