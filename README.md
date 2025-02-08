# Tesla BLE to Mqtt

<img src="./docs/logo.png" width="300">

## Overview

Tesla BLE to Mqtt is a project that bridges Tesla Bluetooth Low Energy (BLE) data to an MQTT broker. This allows you to monitor and interact with your Tesla vehicle using MQTT from Home assistant.
It is a **no compromises solution** with robust, reliable and **fast data updates** via configurable automatic polling, marking sensors as **"Unknown" when vehicle is asleep** or away and does not
keep the **vehicle awake** when polling (but wakes the vehicle up automatically if a command is issued).

Along with keeping the Bluetooth connection alive, it canceles any data polling instantly when a new command is issued, making them **as fast as possible** to execute. After the command is complete,
it will instantly restart its data polling for fast updates after making changes. If for any reason the application runs in to an error it will display it over a seperate "Last error" entity, making
it easy to see when things go wrong.

> [!TIP]
> This repository contains source code for the application running inside the Homeassistant addon. If you want to install this as an addon check out [TeslaBle2Mqtt-addon](https://github.com/Lenart12/TeslaBle2Mqtt-addon) repository.
>
> 
> [![Open your Home Assistant instance and show the add add-on repository dialog with a specific repository URL pre-filled.](https://my.home-assistant.io/badges/supervisor_add_addon_repository.svg)](https://my.home-assistant.io/redirect/supervisor_add_addon_repository/?repository_url=https://github.com/Lenart12/TeslaBle2Mqtt-addon)

> [!CAUTION]
> This project is still in its early stages, so expect some things to change.

## Features

- Connects to Tesla vehicle via TeslaBleHttpProxy
- Publishes vehicle data to MQTT topics
- Supports multiple Tesla vehicles
- Easy command line configuration
- Configurable polling
- Automatic integration with home assistant Mqtt autodiscovery

## Screenshots

<p>
  <img src="https://github.com/user-attachments/assets/6870823b-899b-4706-bfb8-272f8deb32f6" align="top" style="width: 30%">
  <img src="https://github.com/user-attachments/assets/66841ccf-9ed1-446f-adef-f274f25d983e" align="top" style="width: 30%">
  <img src="https://github.com/user-attachments/assets/1e257de9-1b73-4436-a76f-a2cab549910c" align="top" style="width: 30%">
</p>

## Requirements

- Go 1.23
- MQTT broker (e.g., Mosquitto)
- TeslaBleHttpProxy
- Home assistant with mqtt integration (Optional if you do not care about displaying values)

> [!WARNING]
> Until [wimaha/TeslaBleHttpProxy#95](https://github.com/wimaha/TeslaBleHttpProxy/pull/95) gets merged, it is expected to
> use my fork of TeslaBleHttpProxy, which you can get [here](https://github.com/Lenart12/TeslaBleHttpProxy) and then build it
> manually as I can't create releases.

## Installation

### Option 1: Home Assistant Addon (Recommended)

The easiest way to install TeslaBle2Mqtt is through the Home Assistant addon store:

[![Open your Home Assistant instance and show the add add-on repository dialog with a specific repository URL pre-filled.](https://my.home-assistant.io/badges/supervisor_add_addon_repository.svg)](https://my.home-assistant.io/redirect/supervisor_add_addon_repository/?repository_url=https://github.com/Lenart12/TeslaBle2Mqtt-addon)

1. Click the button above to add the repository
2. Find and install "Tesla BLE to MQTT" in the addon store
3. Configure your vehicle VIN and MQTT settings
4. Start the addon

### Option 2: Docker Compose

Pre-built multi-architecture images are available from GitHub Container Registry:

```bash
docker pull ghcr.io/lenart12/teslable2mqtt:latest
```

1. Create a docker-compose.yml file:
```yaml
services:
  teslable2mqtt:
    image: ghcr.io/lenart12/teslable2mqtt:latest
    container_name: teslable2mqtt
    restart: unless-stopped
    network_mode: host
    command:
      - "--proxy-host=http://localhost:8080"
      - "--mqtt-host=localhost"
      - "--mqtt-user=your_username"
      - "--mqtt-pass=your_password"
      - "--vin=YOUR_TESLA_VIN"
      - "--log-level=info"
      - "--force-ansi-color"
```

2. Start the container:
```bash
docker compose up -d
```

> [!NOTE]
> When running in Docker, use `network_mode: host` to allow the container to access services running on the host machine.
> This allows you to use `localhost` to connect to the proxy and MQTT broker running on the host.
> 
> Alternatively, you can use:
> - `host.docker.internal` instead of `localhost` on Windows/macOS
> - The host machine's IP address (e.g., 192.168.1.x)

### Option 3: Manual Build from Source

1. Clone the repository:
    ```sh
    git clone https://github.com/yourusername/TeslaBle2Mqtt.git
    cd TeslaBle2Mqtt
    ```

2. Build and run:
    ```sh
    go build
    ./TeslaBle2Mqtt \
      --proxy-host=http://localhost:8080 \
      --mqtt-host=localhost \
      --mqtt-user=your_username \
      --mqtt-pass=your_password \
      --vin=YOUR_TESLA_VIN \
      --log-level=info
    ```

### Option 4: Manual Docker

1. Build the image:
```bash
docker build -t teslable2mqtt .
```

2. Run the container:
```bash
docker run -d \
  --name teslable2mqtt \
  --network host \
  --restart unless-stopped \
  teslable2mqtt \
  --proxy-host=http://localhost:8080 \
  --mqtt-host=localhost \
  --mqtt-user=your_username \
  --mqtt-pass=your_password \
  --vin=YOUR_TESLA_VIN \
  --log-level=info
```

## Usage

Start the application:
```
$ ./TeslaBle2Mqtt --help
usage: Tesla BLE to Mqtt [-h|--help] -v|--vin "<value>" [-v|--vin "<value>"
                         ...] [-p|--proxy-host "<value>"] [-i|--poll-interval
                         <integer>] [-I|--poll-interval-charging <integer>]
                         [-A|--max-charging-amps <integer>] [-H|--mqtt-host
                         "<value>"] [-P|--mqtt-port <integer>] [-u|--mqtt-user
                         "<value>"] [-w|--mqtt-pass "<value>"] [-q|--mqtt-qos
                         <integer>] [-d|--discovery-prefix "<value>"]
                         [-m|--mqtt-prefix "<value>"] [-y|--sensors-yaml
                         "<value>"] [-r|--reset-discovery] [-l|--log-level
                         "<value>"] [-D|--mqtt-debug] [-V|--reported-version
                         "<value>"] [-C|--reported-config-url "<value>"]
                         [-a|--force-ansi-color] [-L|--log-prefix "<value>"]

                         Expose Tesla sensors and controls to MQTT with Home
                         Assistant discovery

Arguments:

  -h  --help                    Print help information
  -v  --vin                     VIN of the Tesla vehicle (Can be specified
                                multiple times)
  -p  --proxy-host              Proxy host. Default: http://localhost:8080
  -i  --poll-interval           Poll interval in seconds. Default: 90
  -I  --poll-interval-charging  Poll interval in seconds when charging.
                                Default: 20
  -A  --max-charging-amps       Max charging amps. Default: 16
  -H  --mqtt-host               MQTT host. Default: localhost
  -P  --mqtt-port               MQTT port. Default: 1883
  -u  --mqtt-user               MQTT username
  -w  --mqtt-pass               MQTT password
  -q  --mqtt-qos                MQTT QoS. Default: 0
  -d  --discovery-prefix        MQTT discovery prefix. Default: homeassistant
  -m  --mqtt-prefix             MQTT prefix. Default: tb2m
  -y  --sensors-yaml            Path to custom sensors YAML file. Default: 
  -r  --reset-discovery         Reset MQTT discovery
  -l  --log-level               Log level. Default: INFO
  -D  --mqtt-debug              Enable MQTT debug output (sam log level as
                                --log-level)
  -V  --reported-version        Version of this application, reported via Mqtt.
                                Default: dev
  -C  --reported-config-url     URL to the configuration page of this
                                application, reported via Mqtt. Default:
                                {proxy-host}/dashboard
  -a  --force-ansi-color        Force ANSI color output
  -L  --log-prefix              Log prefix. Default: 

```


## Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

