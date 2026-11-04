# HTTP API

All operations must be sent to the port defined in config.yaml using HTTP (no HTTPS). All operations are exclusively JSON request/response based, and must include appropriate headers:
```
Accept: application/json
Content-Type: application/json
```

The accept header may be omitted, but the content-type header is required for all POST operations.

## Status
The status operation reports on the operational status of the server and the status of each light

```
GET /
Accept: application/json

{
  "server": {
    "mode": "auto",
    "attributes": {
      "scenes": "default,patio_sunset_dynamic",
      "calendar_scenes": [
        "default",
        "patio_sunset_dynamic"
      ],
      "additional_scenes": []
    }
  },
  "lights": {
    "front_flood_cl": {
      "available": true,
      "attributes": {
        "programmed_color_name": "warm_white",
        "programmed_color": "W2700/100",
        "actual_color_name": "warm_white",
        "actual_color": "W2700/100",
        "affected_by_scene": true,
        "active_scene": "default",
        "locked": false
      }
    },
    ...
  }
}
```

## Set scenes for the remainder of the day
The day operation will sets one or more additional scenes to be applied for the remainder of the calendar day, or clear/modify scenes previously set with this method.
These scenes are applied in order after all calendar based scenes have been applied.

```
POST /day
Accept: application/json
Content-Type: application/json

{
    "scenes": [ "first_additional_scene", "next_additional_scene", ...]
}

{ 
    "success":true
}
```

## Set lights to explicit color
The status operation sets a target (light or light group) to an explicit color. The color specified may be the name of a color defined in the YAML file, or a raw color specification.

The for value specifies a number of seconds that the light will be locked to this color. Unless unlocked via API, the light will remain on the specified color for this period. After this period expires (or immediately if for is not specified or zero), the light will revert to being controlled by any scene that it currently programming a color for it.

```
POST /set
Accept: application/json
Content-Type: application/json

{
    "target": "office_lights",
    "color": "#00FF00",
    "for": 20
}

{
  "office_ceiling_1": {
    "success": true
  },
  "office_stand_light_upper": {
    "success": false,
    "reason": "Light not available"
  },
  ...
}
```

## Unlock previous locked lights
The unlock operation unlocks lights which may have been locked by a previous set operation. As with set, it targets either a single light or group of lights:

```
POST /unlock
Accept: application/json
Content-Type: application/json

{
    "target": "office_lights"
}

{
    "success":true
}

## Reset operation
The reset operation resets the server to normal operation. Any locked lights are unlocked. Any day scenes are removed.

```
POST /reset
Accept: application/json

.. No body

{
    "success":true
}
```