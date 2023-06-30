package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	upsd_host := flag.String("upsd_host", "localhost", "address of upsd host")
	upsd_port := flag.Int("upsd_port", 3493, "port of upsd server")
	mqtt_host := flag.String("mqtt_host", "localhost", "address of MQTT server")
	mqtt_port := flag.Int("mqtt_prt", 1883, "port of mqtt server")
	mqtt_user := flag.String("mqtt_user", "nut", "MQTT username")
	mqtt_password := os.Getenv("MQTT_PASSWORD")

	mqtt_topic_base := flag.String("mqtt_topic_base", "nut/", "base topic for MQTT messages")

	flag.Parse()

	// Connect to mqtt
	mqtt_url := fmt.Sprintf("tcp://%s:%d", *mqtt_host, *mqtt_port)
	mqtt_client, err := mqttClientNew(mqtt_url, mqtt_user, &mqtt_password)
	checkErrFatal(err)
	defer mqtt_client.Disconnect(250)

	mqtt_client.topic_base = *mqtt_topic_base

	// Publish test message
	err = mqtt_client.publishMessage("test", "Testing and so forth.")
	checkErrFatal(err)

	// Get the list of UPSes from upsd
	ups := getUPSNames(upsd_host, upsd_port)
	for _, u := range ups {
		fmt.Printf("Found UPS: %v (%v)\n", u.name, u.description)
	}

}

func checkErrFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
