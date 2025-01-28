# Tesla BLE to Mqtt

![image](https://github.com/user-attachments/assets/6870823b-899b-4706-bfb8-272f8deb32f6)
![image](https://github.com/user-attachments/assets/66841ccf-9ed1-446f-adef-f274f25d983e)
![image](https://github.com/user-attachments/assets/1e257de9-1b73-4436-a76f-a2cab549910c)

## Overview

Tesla BLE to Mqtt is a project that bridges Tesla Bluetooth Low Energy (BLE) data to an MQTT broker. This allows you to monitor and interact with your Tesla vehicle using MQTT from Home assistant.

> [!CAUTION]
> This project is still in its early stages, so expect unstability.

## Features

- Connects to Tesla vehicle via TeslaBleHttpProxy
- Publishes vehicle data to MQTT topics
- Supports multiple Tesla vehicles
- Easy command line configuration
- Automatic integration with home assistant Mqtt autodiscovery

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
  -y  --sensors-yaml            Path to sensors YAML file. Default:
                                mqtt_sensors.yaml
  -r  --reset-discovery         Reset MQTT discovery
  -l  --log-level               Log level. Default: INFO
  -D  --mqtt-debug              Enable MQTT debug output (sam log level as
                                --log-level)
```


## Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

