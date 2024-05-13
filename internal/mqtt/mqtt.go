package mqtt

import (
	"log"
	"strings"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
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

func (c *mqttClient) PublishMessage(msg *MQTTUpdate) error {
	pub_tok := c.c.Publish(c.topic_base+msg.topic, 0, false, msg.content)
	pub_tok.Wait()

	if pub_tok.Error() != nil {
		return error(pub_tok.Error())
	}
	return nil
}

func (c *mqttClient) Disconnect(code uint) {
	c.c.Disconnect(code)
}

// A struct to describe an MQTT update to be made.
type MQTTUpdate struct {
	topic       string
	content     string
	old_content string
}

func (c *mqttClient) ChannelUpdates(ups_chan chan *upsc.UPSInfo, change_chan chan *MQTTUpdate, wg *sync.WaitGroup) {
	defer wg.Done()
	last_ups := map[string]*upsc.UPSInfo{}
	for {
		ups := <-ups_chan
		old, present := last_ups[ups.Name()]
		if present {
			changed_vars := upsc.GetVarDiff(old, ups)
			if len(changed_vars) > 0 {
				for k, v := range changed_vars {
					topic := ups.Name() + "/" + k
					// upsd uses . and we use / - this isn't always the right way, but it often is, so *shrug*
					topic = strings.Replace(topic, ".", "/", -1)
					change_chan <- &MQTTUpdate{topic, v, old.Vars()[k]}
				}
			}
		} else {
			for k, v := range ups.Vars() {
				topic := ups.Name() + "/" + k
				// upsd uses . and we use / - this isn't always the right way, but it often is, so *shrug*
				topic = strings.Replace(topic, ".", "/", -1)
				change_chan <- &MQTTUpdate{topic, v, ""}
			}
		}
		last_ups[ups.Name()] = ups
	}
}

func (c *mqttClient) ConsumeChannelUpdates(change_chan chan *MQTTUpdate, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		update := <-change_chan
		old := update.old_content
		if old == "" {
			old = "[null]"
		}
		log.Printf("MQTT Change: [%v]\t%v -> %v ", c.GetTopicBase()+update.topic, old, update.content)
		c.PublishMessage(update)
	}
}
