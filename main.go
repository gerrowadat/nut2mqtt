package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	mqtt "github.com/gerrowadat/nut2mqtt/internal/mqtt"
	upsc "github.com/gerrowadat/nut2mqtt/internal/upsc"
)

func main() {
	upsd_host := flag.String("upsd_host", "localhost", "address of upsd host")
	upsd_port := flag.Int("upsd_port", 3493, "port of upsd server")
	mqtt_host := flag.String("mqtt_host", "localhost", "address of MQTT server")
	mqtt_port := flag.Int("mqtt_prt", 1883, "port of mqtt server")
	mqtt_user := flag.String("mqtt_user", "nut", "MQTT username")
	mqtt_password := os.Getenv("MQTT_PASSWORD")

	mqtt_topic_base := flag.String("mqtt_topic_base", "nut/", "base topic for MQTT messages")
	upsd_poll_interval := flag.Int("upsd_poll_interval", 30, "interval between upsd polls")

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

	// Start the upsd polling
	ups_chan := make(chan *upsc.UPSInfo)
	mqtt_change_chan := make(chan *mqtt.MQTTUpdate)

	var wg sync.WaitGroup

	wg.Add(3)

	go upsd_c.WatchForUPSes(time.Duration(*upsd_poll_interval), ups_chan, &wg)
	go mqtt_client.ChannelUpdates(ups_chan, mqtt_change_chan, &wg)
	go mqtt_client.ConsumeChannelUpdates(mqtt_change_chan, &wg)

	wg.Wait()
}

func checkErrFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
