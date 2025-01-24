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
- Automatic integration with home assistant

## Todo

- [x] Project prototype
- [ ] Rewrite the handler to clean up spaghetti code
- [ ] Better availability handling during sleep
- [ ] Home assistant Add-on

## Requirements

- Python 3.7+
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
    python3 -m venv .venv
    source .venv/bin/activate
    pip3 install -r requirements.txt
    ```

## Usage

Start the application:
```
$ python teslable2mqtt.py -h
usage: teslable2mqtt.py [-h] -v VIN -p PROXY_HOST [-i POLL_INTERVAL] [-I POLL_INTERVAL_CHARGING] [-H MQTT_HOST] [-P MQTT_PORT] [-u MQTT_USERNAME] [-w MQTT_PASSWORD] [-d DISCOVERY_PREFIX] [-m MQTT_PREFIX] [-y SENSORS_YAML] [-r] [-l LOG_LEVEL]

Tesla BLE to MQTT proxy

options:
  -h, --help            show this help message and exit
  -v VIN, --vin VIN     VIN of the Tesla vehicle (can be specified multiple times) (default: None)
  -p PROXY_HOST, --proxy-host PROXY_HOST
                        Host of the Tesla BLE proxy (default: None)
  -i POLL_INTERVAL, --poll-interval POLL_INTERVAL
                        Poll interval for vehicle data (default: 90)
  -I POLL_INTERVAL_CHARGING, --poll-interval-charging POLL_INTERVAL_CHARGING
                        Poll interval for vehicle data when car is charging (default: 20)
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


## Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

