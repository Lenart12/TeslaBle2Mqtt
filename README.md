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

- Go
- MQTT broker (e.g., Mosquitto)
- TeslaBleHttpProxy
- Home assistant with mqtt integration

## Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/yourusername/TeslaBle2Mqtt.git
    cd TeslaBle2Mqtt
    ```

2. Install the required Python packages:
    ```sh
    go build .
    ./TeslaBle2Mqtt -h
    ```

## Docker Installation

Pre-built multi-architecture images are available from GitHub Container Registry:

```bash
docker pull ghcr.io/lenart12/teslable2mqtt:latest
```

### Using Docker Compose (recommended)

1. Create a docker-compose.yml file:
```yaml
version: '3.8'

services:
  tesla-ble2mqtt:
    image: ghcr.io/lenart12/teslable2mqtt:latest
    build: .  # Fallback to local build if needed
    container_name: teslable2mqtt
    command:
      - "--mqtt-host=localhost"
      - "--mqtt-user=your_username"
      - "--mqtt-pass=your_password"
      - "--vin=YOUR_TESLA_VIN"
    network_mode: host
    privileged: true
```

### Using Docker manually

1. Build the image:
```bash
docker build -t teslable2mqtt .
```

2. Run the container:
```bash
docker run -d \
  --name tesla-ble2mqtt \
  --network host \
  --privileged \
  tesla-ble2mqtt \
  --mqtt-host localhost \
  --mqtt-user your_username \
  --mqtt-pass your_password \
  --vin YOUR_TESLA_VIN
```

## Usage

Start the application:
```
$ ./TeslaBle2Mqtt --help
usage: Tesla BLE to Mqtt [-h|--help] -v|--vin "<value>" [-v|--vin "<value>"
                         ...] [-p|--proxy-host "<value>"] [-i|--poll-interval
                         <integer>] [-I|--poll-interval-charging <integer>]
                         [-H|--mqtt-host "<value>"] [-P|--mqtt-port <integer>]
                         [-u|--mqtt-user "<value>"] [-w|--mqtt-pass "<value>"]
                         [-q|--mqtt-qos <integer>] [-d|--discovery-prefix
                         "<value>"] [-m|--mqtt-prefix "<value>"]
                         [-y|--sensors-yaml "<value>"] [-r|--reset-discovery]
                         [-l|--log-level "<value>"] [-D|--mqtt-debug]
                         [-V|--reported-version "<value>"]
                         [-C|--reported-config-url "<value>"]

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
```


## Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

