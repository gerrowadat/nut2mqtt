package main

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
)

type mqttClient struct {
	c          mqtt.Client
	topic_base string
}

func mqttClientNew(mqtt_url string, user *string, pass *string) (mqttClient, error) {

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

func (c *mqttClient) publishMessage(topic string, content string) error {
	log.Print("Publish: " + c.topic_base + topic + " => " + content)
	pub_tok := c.c.Publish(c.topic_base+topic, 0, false, content)
	pub_tok.Wait()

	if pub_tok.Error() != nil {
		return error(pub_tok.Error())
	}
	return nil
}

func (c *mqttClient) Disconnect(code uint) {
	c.c.Disconnect(code)
}
