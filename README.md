# nut2mqtt
Shunt NUT server (UPS) info to MQTT 

NOTE: This software is pre-alpha, I'm still getting even the basics working.

Usage
=====

Get nut (`apt install nut`) working on your machine. Verify with telnet to port 3493 and type `LIST UPS`.

Run the thing:

```
MQTT_PASSWORD=mymqttpassword go run *.go --mqtt_host=my.mqtt.host --upsd_host=my.mqtt.host --mqtt_user=mymqttuser
```

This will capture variables from all connected and configured UPSes and populate a path in the mqtt namespace with
simple k/v pairs - see the output of the above LIST UPS command for an idea of what gets updated.

Use `--mqtt_topic_base` to specify where in the mqtt namespace all this goes. We then populate:

```
base/lastupdate = &lt;timestamp&gt;
base/upsname/battery/charge = 100
base/upsname/... = etc...
```
