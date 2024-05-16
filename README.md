# nut2mqtt
Shunt NUT server (UPS) info to MQTT 

Usage
=====

Get nut (`apt install nut`) working on your machine. Verify with telnet to port 3493 and type `LIST UPS`. If you're connecting remotely, you need to tell it to run on 0.0.0.0 and so on.

Run the thing:

```
MQTT_PASSWORD=mymqttpassword nut2mqtt --mqtt-host=my.mqtt.host --upsd-hosts=upshost1,upshost2 --mqtt-user=mymqttuser --mqtt-topic-base=nut2mqtt/
```

This will capture variables from all connected and configured UPSes on each of the specified hosts, and populate a path in the mqtt namespace with simple k/v pairs - see the output of the above LIST UPS command for an idea of what gets updated.

Use `--mqtt-topic-base` to specify where in the mqtt namespace all this goes.

We then populate:

```
base/bridge/state = online|offline
base/hosts/upshost1/upsname/battery/charge = 100
...etc...
```

Grab a utility like MQTT explorer to see what else gets populated.
