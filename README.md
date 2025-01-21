# Tesla BLE to Mqtt

## Overview

Tesla BLE to Mqtt is a project that bridges Tesla Bluetooth Low Energy (BLE) data to an MQTT broker. This allows you to monitor and interact with your Tesla vehicle using MQTT from Home assistant.

## Features

- Connects to Tesla vehicle via TeslaBleHttpProxy
- Publishes vehicle data to MQTT topics
- Supports multiple Tesla vehicles
- Easy command line configuration
- Automatic integration with home assistant

## Requirements

- Python 3.7+
- MQTT broker (e.g., Mosquitto)
- TeslaBleHttpProxy

## Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/yourusername/TeslaBle2Mqtt.git
    cd TeslaBle2Mqtt
    ```

2. Install the required Python packages:
    ```sh
    python3 -m venv .venv
    source .venv/bin/activate
    pip3 install -r requirements.txt
    ```

## Usage

Start the application:
```
$ python3 teslable2mqtt.py --help
usage: teslable2mqtt.py [-h] -v VIN -p PROXY_HOST [-H MQTT_HOST] [-P MQTT_PORT] [-u MQTT_USERNAME] [-w MQTT_PASSWORD] [-d DISCOVERY_PREFIX]
                        [-m MQTT_PREFIX] [-y SENSORS_YAML] [-r] [-l LOG_LEVEL]

Tesla BLE to MQTT proxy

options:
  -h, --help            show this help message and exit
  -v VIN, --vin VIN     VIN of the Tesla vehicle (can be specified multiple times) (default: None)
  -p PROXY_HOST, --proxy-host PROXY_HOST
                        Host of the Tesla BLE proxy (default: None)
  -H MQTT_HOST, --mqtt-host MQTT_HOST
                        MQTT host (default: homeassistant)
  -P MQTT_PORT, --mqtt-port MQTT_PORT
                        MQTT port (default: 1883)
  -u MQTT_USERNAME, --mqtt-username MQTT_USERNAME
                        MQTT username (default: None)
  -w MQTT_PASSWORD, --mqtt-password MQTT_PASSWORD
                        MQTT password (default: None)
  -d DISCOVERY_PREFIX, --discovery-prefix DISCOVERY_PREFIX
                        Home Assistant MQTT Discovery prefix (default: homeassistant)
  -m MQTT_PREFIX, --mqtt-prefix MQTT_PREFIX
                        MQTT prefix (default: tb2m)
  -y SENSORS_YAML, --sensors-yaml SENSORS_YAML
                        Path to the YAML file with MQTT sensors configuration (default: mqtt_sensors.yaml)
  -r, --reset-discovery
                        Reset MQTT discovery (for development) (default: False)
  -l LOG_LEVEL, --log-level LOG_LEVEL
                        Log level (default: INFO)
```

![image](https://github.com/user-attachments/assets/70bb8b2e-b854-4b89-970b-87ebd084526d)


## Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

