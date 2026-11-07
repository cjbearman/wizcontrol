# API (HTTP and MQTT)
Wizcontrol can provide APIs for status and control using both HTTP and MQTT.

To enable the HTTP API, ensure you include the http section in config.yaml.

To enable the MQTT API, ensure you include the api section in config.yaml.

Both APIs perform the same functions using the same requests and responses, all of which are JSON based.

# Endpoints


## HTTP Service Status
The server status API returns the current state of the service, and all lights.

```
GET /

{
  "server": {
    "mode": "auto",
    "attributes": {
      "scenes": "default,patio_ocean_dynamic",
      "active_calendars": [
        "default",
        "patio_changeup"
      ],
      "calendar_scenes": [
        "default",
        "patio_ocean_dynamic"
      ],
      "additional_scenes": []
    }
  },
  "lights": {
    "front_flood_cl": {
      "available": false,
      "attributes": {
        "ip": "192.168.20.12",
        "on": false,
        "programmed_color_name": "warm_white",
        "programmed_color": "w2700/100",
        "actual_color_name": "unknown",
        "actual_color": "unknown",
        "affected_by_scene": true,
        "active_scene": "default",
        "locked": false
      }
    },
    ...
  }
}
```

## MQTT /wizcontrol/server
Any time the server status changes, the "server" element of the status response above is automatically dispatched (retained) to this topic.

## MQTT /wizcontrol/light/{light}
Any time an individual light status changes, it's object from the above response is automatically dispatched (retained) to this topic, where {light} is the light name.

