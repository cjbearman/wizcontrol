# wizcontrol
A golang based runtime to provide local control of Wiz connected lighting devices.

Features:
* YAML based configuration
* Define scenes for different lights or groups of lights for automatic activation based on calendar definitions.
* HTTP and MQTT APIs for status and control
* Lights configured with IP address or MAC address (using wiz discovery)
* Does not rely on light being continually powered up, lights will be adjusted (if needed) shortly after they power-up and become available on the network.

## Using docker
Create your config.yaml and place that in a suitable directory on your docker host (~/wizcontrol_config used in this example).

Run as follows:

```
docker run -d \
	-p 12120:12120 \
	--user "$(id -u):$(id -g)" \
	--name=wizcontrol \
	--restart=unless-stopped \
	-v ${HOME}/wizcontrol_config:/config \
	cbearman/wizcontrol:latest
```

In the case that your lights are not all on fixed IP addresses and you need to configure lights using MAC addresses, you'll need to use host networking to run your container, so that broadcast network operations can be conducted.

```
docker run -d \
	-p 12120:12120 \
	--user "$(id -u):$(id -g)" \
	--name=wizcontrol \
	--restart=unless-stopped \
	-v ${HOME}/wizcontrol_config:/config \
    -network host \
	cbearman/wizcontrol:latest
```

## Without docker
You will need go (golang) installed on your machine

Install wizcontrol and wizfind:
```
go install github.com/cjbearman/wizcontrol/cmd/wizcontrol@latest
go install github.com/cjbearman/wizcontrol/cmd/wizfind@latest
```

Create your config.yaml, and run as follows:
```
wizcontrol -d -config config.yaml
```

A JSON schema for config.yaml is available in [config.schema.json](./config.schema.json).

