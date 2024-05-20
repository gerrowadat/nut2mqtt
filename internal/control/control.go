// Program Control nonsense.

package control

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	channels "github.com/gerrowadat/nut2mqtt/internal/channels"
	metrics "github.com/gerrowadat/nut2mqtt/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
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
			Control:       make(chan *channels.ControlMessage),
			Ups:           make(chan *channels.UPSInfo),
			Metrics:       make(chan *channels.UPSVariableUpdate),
			MqttConverter: make(chan *channels.UPSVariableUpdate),
			Mqtt:          make(chan *channels.MQTTUpdate),
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

// A decaying cache of UPS info. If we don't see a UPS for a while, we remove it.
type DecayingUPSCacheEntry struct {
	ups       *channels.UPSInfo
	last_seen *time.Time
}

func PruneUPSCache(cache map[string]*DecayingUPSCacheEntry, expiry time.Duration) {
	expiry_time := time.Now().Add(-expiry)
	for k, v := range cache {
		if v.last_seen.Before(expiry_time) {
			log.Printf("Pruning UPS cache entry: %v ", k)
			delete(cache, k)
		}
	}

}

func (c *Controller) EmitVariableUpdate(chg *channels.UPSVariableUpdate) {
	// Send to all the channels that care about this.
	// Remember these are blocking.
	c.mr.Metrics().UPSVariableUpdatesProcessed.Inc()
	c.cb.Metrics <- chg
	c.cb.MqttConverter <- chg
}

func (c *Controller) UPSVariableUpdateMultiplexer() {
	defer c.wg.Done()
	ups_info := map[string]*DecayingUPSCacheEntry{}
	for {
		// Get a UPSInfo from the channel
		u := <-c.cb.Ups
		// Prune our UPS cache first
		PruneUPSCache(ups_info, 5*time.Minute)
		// If this is a brand new UPS, we need to emit all of its variables.
		_, present := ups_info[u.Name]
		if !present {
			for k, v := range u.Vars {
				c.EmitVariableUpdate(&channels.UPSVariableUpdate{Host: u.Host, UpsName: u.Name, VarName: k, Content: v, OldContent: ""})
			}
		} else {
			// This is an existing UPS. We need to diff the variables.
			old := ups_info[u.Name].ups
			for k, v := range u.Vars {
				if old.Vars[k] != v {
					c.EmitVariableUpdate(&channels.UPSVariableUpdate{Host: u.Host, UpsName: u.Name, VarName: k, Content: v, OldContent: old.Vars[k]})
				}
			}
		}
		// Plop this into the cache ragardless.
		ups_info[u.Name] = &DecayingUPSCacheEntry{ups: u, last_seen: &time.Time{}}
	}
}

func (c *Controller) MetricsUpdateConsumer() {
	defer c.wg.Done()
	for {
		up := <-c.cb.Metrics
		for _, m := range metrics.UPSMetricsList {
			if m.NutVariable == up.VarName {
				// Emit this metric with the given host and ups.
				if m.Type == "gaugevec" {
					met := c.mr.GetUPSMetric(m.Name).(*prometheus.GaugeVec)
					val, err := strconv.ParseFloat(up.Content, 64)
					if err != nil {
						log.Printf("Error parsing float for UPS variable %v: %v", m.Name, err)
						continue
					}
					met.With(prometheus.Labels{"host": up.Host, "ups": up.UpsName}).Set(val)
				}
			}
		}
	}
}
