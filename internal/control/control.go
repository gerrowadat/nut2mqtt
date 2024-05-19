// Program Control nonsense.

package control

import (
	"fmt"
	"sync"

	channels "github.com/gerrowadat/nut2mqtt/internal/channels"
	"github.com/gerrowadat/nut2mqtt/internal/metrics"
)

type Controller struct {
	cb         *channels.ChannelBundle
	mr         *metrics.MetricRegistry
	wg         *sync.WaitGroup
	mqtt_topic string
}

func NewController(mqtt_topic string) Controller {
	var wg sync.WaitGroup
	// Set to 1, as we want to exit if even 1 subprocess dies.
	wg.Add(1)
	return Controller{
		cb: &channels.ChannelBundle{
			Control: make(chan *channels.ControlMessage),
			Ups:     make(chan *channels.UPSInfo),
			Mqtt:    make(chan *channels.MQTTUpdate),
		},
		mr:         metrics.NewMetricRegistry(),
		wg:         &wg,
		mqtt_topic: mqtt_topic}
}

func (c Controller) Startup(comment string, args ...interface{}) {
	comment = fmt.Sprintf("Startup: "+comment, args...)
	c.cb.Control <- &channels.ControlMessage{Operation: "startup", Comment: comment}
}

func (c Controller) Shutdown(comment string, args ...interface{}) {
	comment = fmt.Sprintf("Shutdown: "+comment, args...)
	c.cb.Control <- &channels.ControlMessage{Operation: "shutdown", Comment: comment}
}

// Redirections to other bits, I am a bad programmer man.
func (c Controller) WaitGroupDone() {
	c.wg.Done()
}

func (c Controller) Wait() {
	c.wg.Wait()
}

func (c Controller) MetricRegistry() *metrics.MetricRegistry {
	return c.mr
}

func (c Controller) Channels() *channels.ChannelBundle {
	return c.cb
}

func (c *Controller) ControlMessageConsumer() {
	defer c.wg.Done()
	for {
		msg := <-c.cb.Control
		c.mr.Metrics().ControlMessagesProcessed.Inc()
		fmt.Println("Processing Control message: ", msg.String())
		switch msg.Operation {
		case "startup":
			c.cb.Mqtt <- &channels.MQTTUpdate{Topic: c.mqtt_topic + "/state", Content: "online"}
		case "shutdown":
			c.cb.Mqtt <- &channels.MQTTUpdate{Topic: c.mqtt_topic + "/state", Content: "offline"}
			// returning will exit the consumer, and process will end.
			return
		default:
			fmt.Println("Unknown operation on control channel: ", msg.Operation)
		}
	}
}
