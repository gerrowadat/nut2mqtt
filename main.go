package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	channels "github.com/gerrowadat/nut2mqtt/internal/channels"
	control "github.com/gerrowadat/nut2mqtt/internal/control"
	http "github.com/gerrowadat/nut2mqtt/internal/http"
	mqtt "github.com/gerrowadat/nut2mqtt/internal/mqtt"
	upsc "github.com/gerrowadat/nut2mqtt/internal/upsc"
)

func main() {
	upsd_hosts := flag.String("upsd-hosts", "localhost", "address of upsd host(s), comma-separated")
	upsd_port := flag.Int("upsd-port", 3493, "port of upsd server")
	mqtt_host := flag.String("mqtt-host", "localhost", "address of MQTT server")
	mqtt_port := flag.Int("mqtt-port", 1883, "port of mqtt server")
	mqtt_user := flag.String("mqtt-user", "nut", "MQTT username")
	mqtt_password := os.Getenv("MQTT_PASSWORD")

	mqtt_topic_base := flag.String("mqtt-topic-base", "nut/", "base topic for MQTT messages")
	upsd_poll_interval := flag.Int("upsd-poll-interval", 30, "interval between upsd polls")
	upsd_cache_lifetime := flag.String("upsd-cache-lifetime", "60s", "lifetime of upsd cache entries")

	control_topic := flag.String("control-topic", "bridge", "subtopic for control/alive messages")

	http_listen := flag.String("http-listen", ":8080", "Where the http server should listen (default :8080)")

	flag.Parse()

	// Get the list of UPSes from upsd
	ups_hosts := upsc.NewUPSHosts(*upsd_hosts, *upsd_port)

	// Connect to mqtt
	mqtt_url := fmt.Sprintf("tcp://%s:%d", *mqtt_host, *mqtt_port)
	mqtt_client, err := mqtt.NewMQTTClient(mqtt_url, mqtt_user, &mqtt_password)
	if err != nil {
		log.Fatal("MQTT fatal error: ", err)
	}
	defer mqtt_client.Disconnect(250)
	mqtt_client.SetTopicBase(*mqtt_topic_base)

	// Create the controller
	upsd_cache_lifetime_duration, err := time.ParseDuration(*upsd_cache_lifetime)
	if err != nil {
		log.Fatal("Could not parse --upsd-cache-lifetime: ", err)
	}
	controller := control.NewController(*control_topic, upsd_cache_lifetime_duration)

	// Consume control messages, startup, shutdown, etc.
	go controller.ControlMessageConsumer()

	// Produce UPS info by talking to nut instances.
	go ups_hosts.UPSInfoProducer(&controller, time.Duration(*upsd_poll_interval))

	// Multiplex these UPS changes into UPSVariableUpdate messages
	go controller.UPSVariableUpdateMultiplexer()

	// Produce MQTT updates from UPSVariableUpdate messages
	go mqtt_client.UpdateProducer(&controller)

	// Consume UPS changes and Do the Needful
	go mqtt_client.UpdateConsumer(&controller)
	go controller.MetricsUpdateConsumer()

	// Start the http server
	go http.HTTPServer(&controller, http_listen)

	controller.Startup("Online at %v", time.Now().String())

	controller.Wait()

	// One of our goroutines has died, send our offline message and exit.
	mqtt_client.PublishMessage(&channels.MQTTUpdate{Topic: *control_topic, Content: "offline"})
}
