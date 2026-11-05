# wizfind

Wizfind is a helper that can be used to discover your wiz lights

## Discovering your lights

Use the "-f" option to find run wiz discovery to find your lights. Example:

```
wizfind -f -b 192.168.21.255
Discovered device, MAC: 98:77:d5:46:6e:d4, IP: 192.168.20.192
Discovered device, MAC: 98:77:d5:30:8a:14, IP: 192.168.20.244
Discovered device, MAC: 98:77:d5:46:9b:d8, IP: 192.168.20.163
Discovered device, MAC: cc:40:85:c5:a8:de, IP: 192.168.20.13
Discovered device, MAC: 98:77:d5:30:1a:fc, IP: 192.168.20.84
Discovered device, MAC: 98:77:d5:a6:b7:3a, IP: 192.168.20.251
Found 6 devices:
```

Your lights must be powered on and connected to the network to be discoverable.

If no lights are discovered, you might need to explicitly set the broadcast address of the network on which the lights are located. This can be done with the -b option (i.e. -b 192.168.1.255).

Note that discovery is not 100% reliable.

## Changing the color of a light to identify it
Since discovery tells you only the MAC address and IP address of the light, there is no way that discovery can tell you wish light is which.

The easiest way to proceed is to individual lights to a specific color, to identify them, for example:

```
wizfind -s 192.168.20.192 -c '#00FF00'
Setting light at "192.168.20.192" to "#00FF00"
Done.
```
You can specify any valid color string in the -c option, or omit this option for "red".

## What to do once you've discovered and identified your lights
Once you've found the IPs and MACs for all your lights and figured out which is which, your next step should be to go to your home router and set up a DHCP reservation for each light such that the MAC is permanently tied to the IP.