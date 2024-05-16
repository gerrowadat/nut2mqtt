package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	channels "github.com/gerrowadat/nut2mqtt/internal/channels"
	control "github.com/gerrowadat/nut2mqtt/internal/control"
	mqtt "github.com/gerrowadat/nut2mqtt/internal/mqtt"
	upsc "github.com/gerrowadat/nut2mqtt/internal/upsc"
)

func main() {
	upsd_host := flag.String("upsd-host", "localhost", "address of upsd host")
	upsd_port := flag.Int("upsd-port", 3493, "port of upsd server")
	mqtt_host := flag.String("mqtt-host", "localhost", "address of MQTT server")
	mqtt_port := flag.Int("mqtt-prt", 1883, "port of mqtt server")
	mqtt_user := flag.String("mqtt-user", "nut", "MQTT username")
	mqtt_password := os.Getenv("MQTT_PASSWORD")

	mqtt_topic_base := flag.String("mqtt-topic-base", "nut/", "base topic for MQTT messages")
	upsd_poll_interval := flag.Int("upsd-poll-interval", 30, "interval between upsd polls")

	control_topic := flag.String("control-topic", "bridge", "subtopic for control/alive messages")

	flag.Parse()

	// Get the list of UPSes from upsd
	upsd_c := upsc.NewUPSDClient(*upsd_host, *upsd_port)

	// Connect to mqtt
	mqtt_url := fmt.Sprintf("tcp://%s:%d", *mqtt_host, *mqtt_port)
	mqtt_client, err := mqtt.NewMQTTClient(mqtt_url, mqtt_user, &mqtt_password)
	if err != nil {
		log.Fatal("MQTT fatal error: ", err)
	}
	checkErrFatal(err)
	defer mqtt_client.Disconnect(250)
	mqtt_client.SetTopicBase(*mqtt_topic_base)

	// Various channels for processing
	ups_chan := make(chan *channels.UPSInfo)
	mqtt_change_chan := make(chan *channels.MQTTUpdate)

	// Create the controller
	controller := control.NewController(mqtt_change_chan, *control_topic)

	var wg sync.WaitGroup

	// This is 1 because we want to exit if any of the below goroutines exit.
	wg.Add(1)

	go controller.ControlMessageConsumer(&wg)
	go upsd_c.UPSInfoProducer(&controller, &wg, ups_chan, time.Duration(*upsd_poll_interval))
	go mqtt_client.UpdateProducer(&controller, ups_chan, mqtt_change_chan, &wg)
	go mqtt_client.UpdateConsumer(&controller, mqtt_change_chan, &wg)

	controller.Startup("Online at %v", time.Now().String())

	wg.Wait()

	// One of our goroutines has died, send our offline message and exit.
	mqtt_client.PublishMessage(&channels.MQTTUpdate{Topic: *control_topic, Content: "offline"})
}

func checkErrFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
