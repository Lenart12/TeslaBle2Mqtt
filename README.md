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
> It is expected to use my fork of TeslaBleHttpProxy, which you can get [here](https://github.com/Lenart12/TeslaBleHttpProxy)
> See [Installation](#option-2-docker-compose) for more information.

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
    command:
      - "--vin=YOUR_TESLA_VIN" # VIN of the Tesla vehicle
      - "--proxy-host=http://teslablehttpproxy:8080" # URL to the TeslaBleHttpProxy
      - "--mqtt-host=your_host" # MQTT host (if localhost, see note below)
      - "--mqtt-user=your_username" # MQTT username
      - "--mqtt-pass=your_password" # MQTT password
      - "--reset-discovery" # Reset MQTT discovery
      # - "--log-level=INFO" # Log level (INFO, DEBUG, WARN, ERROR, FATAL)
      # - "--poll-interval=90" # Poll interval in seconds
      # - "--poll-interval-charging=20" # Poll interval in seconds when charging
      # - "--fast-poll-time=120" # Period in seconds after discover, wakeup or command that polling is done without reduced interval
      # - "--max-charging-amps=16" # Max charging amps
      # - "--mqtt-port=1883" # MQTT port
      # - "--mqtt-qos=0" # MQTT QoS
      # - "--reported-version=dev" # Version of this application, reported via Mqtt
      - "--force-ansi-color" # Force ANSI color output
      # - "--log-prefix=" # Log prefix

  teslablehttpproxy:
    ##################################
    #
    # Read here!!!
    # Uncomment the following line with the version you want to use:
    #
    #   * latest (Exclusive ownership of the BT adapter)
    #     Better if you have a BT adapter and plan to  use it *only* for TeslaBleHttpProxy
    #     # If you use this version, you also need to add certain privileges to the container:
    #     network_mode: host,
    #     cap_add: NET_ADMIN
    #
    #   * latest-bluez (Using the Linux BlueZ stack)
    #     If you plan to use the same adapter for other applications.
    #     If you use this version, you also need to mount dbus socket:
    #     volumes:
    #      - /var/run/dbus:/var/run/dbus
    #
    #     You can switch between the two versions if you are experiencing issues with one of them.
    #
    # For more information see raw HCI section here:
    #     https://github.com/Lenart12/TeslaBle2Mqtt-addon/blob/main/TeslaBle2Mqtt/DOCS.md#raw-hci
    #
    ##################################

    # image: ghcr.io/lenart12/teslablehttpproxy:latest
    # image: ghcr.io/lenart12/teslablehttpproxy:latest-bluez
    container_name: teslablehttpproxy
    restart: unless-stopped
    command:
      - "--keys=/key" # Path to public and private keys, /key is default mount point
      # - "--logLevel=INFO" # DEBUG, INFO, WARN, ERROR, FATAL
      # - "--httpListenAddress=:8080"
      # - "--scanTimeout=1" # Time in seconds to scan for BLE beacons during device scan (0 = max)
      # - "--cacheMaxAge=5" # Time in seconds for Cache-Control header (0 = no cache)
      # - "--btAdapter=hci0" # Bluetooth adapter ID to use ("hciX"). Default: hci0
    ports:
      - "8080:8080"

    # Uncomment the following lines if you are using "latest" version
    # cap_add:
    #   - NET_ADMIN
    # network_mode: host

    volumes:
      - /path/to/your/keys:/key
      # - Uncomment the following line if you are using "latest-bluez" version
      # - /var/run/dbus:/var/run/dbus
```

2. Start the containers:
```bash
docker compose up -d
```

> [!NOTE]
> When running in Docker, use `network_mode: host` to allow the container to access services running on the host machine.
> This allows you to use `localhost` to connect to MQTT broker running on the host.
> 
> Alternatively, you can use:
> - `host.docker.internal` instead of `localhost` for MQTT on Windows/macOS
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
                         [-f|--fast-poll-time <integer>]
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
  -f  --fast-poll-time          Period in seconds after discover, wakeup or
                                command that polling is done without reduced
                                interval. Default: 120
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

