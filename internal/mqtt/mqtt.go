package mqtt

import (
	"fmt"
	"log"
	"strings"
	"sync"

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

func (c *mqttClient) SetTopicBase(topic_base string) {
	c.topic_base = topic_base
}

func (c *mqttClient) GetTopicBase() string {
	return c.topic_base
}

func (c *mqttClient) PublishMessage(msg *channels.MQTTUpdate) error {
	pub_tok := c.c.Publish(c.topic_base+msg.Topic, 0, false, msg.Content)
	pub_tok.Wait()

	if pub_tok.Error() != nil {
		return error(pub_tok.Error())
	}
	return nil
}

func (c *mqttClient) Subscribe(topic string, callback mqtt.MessageHandler) error {
	token := c.c.Subscribe(topic, 0, callback)
	token.Wait()
	if token.Error() != nil {
		return error(token.Error())
	}
	return nil
}

func (c *mqttClient) Disconnect(code uint) {
	c.c.Disconnect(code)
}

func (c *mqttClient) UpdateProducer(controller *control.Controller, ups_chan chan *channels.UPSInfo, change_chan chan *channels.MQTTUpdate, wg *sync.WaitGroup) {
	defer wg.Done()
	last_ups := map[string]*channels.UPSInfo{}
	for {
		ups := <-ups_chan
		old, present := last_ups[ups.Name]
		if present {
			changed_vars := upsc.GetVarDiff(old, ups)
			if len(changed_vars) > 0 {
				for k, v := range changed_vars {
					topic := fmt.Sprintf("hosts/%v/%v/%v", ups.Host, ups.Name, k)
					// upsd uses . and we use / - this isn't always the right way, but it often is, so *shrug*
					topic = strings.Replace(topic, ".", "/", -1)
					change_chan <- &channels.MQTTUpdate{Topic: topic, Content: v, OldContent: old.Vars[k]}
				}
			}
		} else {
			for k, v := range ups.Vars {
				topic := fmt.Sprintf("hosts/%v/%v/%v", ups.Host, ups.Name, k)
				// upsd uses . and we use / - this isn't always the right way, but it often is, so *shrug*
				topic = strings.Replace(topic, ".", "/", -1)
				change_chan <- &channels.MQTTUpdate{Topic: topic, Content: v, OldContent: ""}
			}
		}
		last_ups[ups.Name] = ups
	}
}

func (c *mqttClient) UpdateConsumer(controller *control.Controller, change_chan chan *channels.MQTTUpdate, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		update := <-change_chan
		old := update.OldContent
		if old == "" {
			old = "[null]"
		}
		log.Printf("MQTT Change: [%v]\t%v -> %v ", c.GetTopicBase()+update.Topic, old, update.Content)
		c.PublishMessage(update)
	}
}
