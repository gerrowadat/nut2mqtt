package mqtt

import (
	"fmt"
	"log"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	channels "github.com/gerrowadat/nut2mqtt/internal/channels"
	control "github.com/gerrowadat/nut2mqtt/internal/control"
	"github.com/gerrowadat/nut2mqtt/internal/upsc"
)

type mqttClient struct {
	c          mqtt.Client
	topic_base string
}

func NewMQTTClient(mqtt_url string, user *string, pass *string) (mqttClient, error) {

	ret := mqttClient{}

	log.Print("Connecting to MQTT at " + mqtt_url + " as " + *user)
	opts := mqtt.NewClientOptions()
	opts.AddBroker(mqtt_url)
	opts.SetClientID("nut2mqtt")
	opts.SetUsername(*user)
	opts.SetPassword(*pass)
	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return ret, error(token.Error())
	}

	ret.c = client

	return ret, nil
}

func (m *mqttClient) SetTopicBase(topic_base string) {
	m.topic_base = topic_base
}

func (m *mqttClient) GetTopicBase() string {
	return m.topic_base
}

func (m *mqttClient) PublishMessage(msg *channels.MQTTUpdate) error {
	pub_tok := m.c.Publish(m.topic_base+msg.Topic, 0, false, msg.Content)
	pub_tok.Wait()

	if pub_tok.Error() != nil {
		return error(pub_tok.Error())
	}
	return nil
}

func (m *mqttClient) Disconnect(code uint) {
	m.c.Disconnect(code)
}

func (m *mqttClient) UpdateProducer(c *control.Controller) {
	defer c.WaitGroupDone()
	last_ups := map[string]*channels.UPSInfo{}
	for {
		ups := <-c.Channels().Ups
		old, present := last_ups[ups.Name]
		if present {
			changed_vars := upsc.GetVarDiff(old, ups)
			if len(changed_vars) > 0 {
				for k, v := range changed_vars {
					topic := fmt.Sprintf("hosts/%v/%v/%v", ups.Host, ups.Name, k)
					// upsd uses . and we use / - this isn't always the right way, but it often is, so *shrug*
					topic = strings.Replace(topic, ".", "/", -1)
					c.Channels().Mqtt <- &channels.MQTTUpdate{Topic: topic, Content: v, OldContent: old.Vars[k]}
				}
			}
		} else {
			for k, v := range ups.Vars {
				topic := fmt.Sprintf("hosts/%v/%v/%v", ups.Host, ups.Name, k)
				// upsd uses . and we use / - this isn't always the right way, but it often is, so *shrug*
				topic = strings.Replace(topic, ".", "/", -1)
				c.Channels().Mqtt <- &channels.MQTTUpdate{Topic: topic, Content: v, OldContent: ""}
			}
		}
		last_ups[ups.Name] = ups
	}
}

func (m *mqttClient) UpdateConsumer(c *control.Controller) {
	defer c.WaitGroupDone()
	for {
		update := <-c.Channels().Mqtt
		old := update.OldContent
		if old == "" {
			old = "[null]"
		}
		log.Printf("MQTT Change: [%v]\t%v -> %v ", m.GetTopicBase()+update.Topic, old, update.Content)
		c.MetricRegistry().Metrics().MQTTUpdatesProcessed.Inc()
		m.PublishMessage(update)
	}
}
