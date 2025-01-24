import asyncio
import argparse
import os
import sys
import logging
import threading
import time
from typing import List, Optional
from pydantic import BaseModel
from paho.mqtt.client import Client
import urllib

import requests
from mqtt_ha import *

logging.basicConfig(level=logging.INFO, format='%(asctime)s:%(levelname)s - %(message)s')
log = logging.getLogger()

class Settings(BaseModel):
    log_level: Optional[str] = "INFO"
    vins: Optional[List[str]] = []
    proxy_host: Optional[str] = "http://localhost:8080"
    poll_interval: Optional[int] = 300
    poll_interval_charging: Optional[int] = 20
    mqtt_host: Optional[str] = "homeassistant"
    mqtt_port: Optional[int] = 1883
    mqtt_username: Optional[str] = None
    mqtt_password: Optional[str] = None
    discovery_prefix: Optional[str] = "homeassistant"
    mqtt_prefix: Optional[str] = "tb2m"
    reset_discovery: Optional[bool] = False
    sensors_yaml: Optional[str] = "mqtt_sensors.yaml"

def parse_args(args) -> Settings:
    defaults = Settings()

    parser = argparse.ArgumentParser(description="Tesla BLE to MQTT proxy", formatter_class=argparse.ArgumentDefaultsHelpFormatter)

    def vin_type(value):
        if len(value) != 17:
            raise argparse.ArgumentTypeError(f"VIN must be 17 characters long (got {len(value)})")
        return value
    parser.add_argument("-v", "--vin", required=True, action='append', type=vin_type, help="VIN of the Tesla vehicle (can be specified multiple times)")
    def url_type(value):
        url = urllib.parse.urlparse(value)
        if not url.scheme or not url.netloc:
            raise argparse.ArgumentTypeError("Invalid URL")
        if url.path != "":
            raise argparse.ArgumentTypeError("URL can't have a path")
        
        return value
    parser.add_argument("-p", "--proxy-host", required=True, help="Host of the Tesla BLE proxy", type=url_type)
    def poll_interval_type(value):
        if not isinstance(value, int):
            raise argparse.ArgumentTypeError("Poll interval must be an integer")
        if value < 1:
            raise argparse.ArgumentTypeError("Poll interval must be at least 1")
        return value
    parser.add_argument("-i", "--poll-interval", default=defaults.poll_interval, help="Poll interval for vehicle data", type=poll_interval_type)
    parser.add_argument("-I", "--poll-interval-charging", default=defaults.poll_interval_charging, help="Poll interval for vehicle data when car is charging", type=poll_interval_type)
    parser.add_argument("-H", "--mqtt-host", default=defaults.mqtt_host, help="MQTT host")
    parser.add_argument("-P", "--mqtt-port", default=defaults.mqtt_port, help="MQTT port", type=int)
    parser.add_argument("-u", "--mqtt-username", help="MQTT username")
    parser.add_argument("-w", "--mqtt-password", help="MQTT password")
    parser.add_argument("-d", "--discovery-prefix", default=defaults.discovery_prefix, help="Home Assistant MQTT Discovery prefix")
    parser.add_argument("-m", "--mqtt-prefix", default=defaults.mqtt_prefix, help="MQTT prefix")
    def yaml_exists(value):
        if not os.path.exists(value):
            raise argparse.ArgumentTypeError(f"File {value} does not exist")
        return value
    parser.add_argument("-y", "--sensors-yaml", default=defaults.sensors_yaml, help="Path to the YAML file with MQTT sensors configuration", type=yaml_exists)
    def log_level_type(value):
        if value not in ["DEBUG", "INFO", "WARNING", "WARN", "ERROR", "FATAL", "CRITICAL"]:
            raise argparse.ArgumentTypeError(f"Invalid log level: {value}")
        return value
    parser.add_argument("-r", "--reset-discovery", default=defaults.reset_discovery, help="Reset MQTT discovery (for development)", action="store_true")
    parser.add_argument("-l", "--log-level", default=defaults.log_level, help="Log level", type=log_level_type)

    args = parser.parse_args(args)
    log.setLevel(args.log_level)

    log.debug(args)

    return Settings(
        vins=args.vin,
        proxy_host=args.proxy_host,
        mqtt_host=args.mqtt_host,
        mqtt_port=args.mqtt_port,
        mqtt_username=args.mqtt_username,
        mqtt_password=args.mqtt_password,
        discovery_prefix=args.discovery_prefix,
        mqtt_prefix=args.mqtt_prefix,
        reset_discovery=args.reset_discovery,
        sensors_yaml=args.sensors_yaml
    )

def init_discovery(settings: Settings) -> dict:
    conf = parse_config(settings.sensors_yaml)
    discoveries, topic_bindings = generate_discovery(settings, conf)
    return discoveries, topic_bindings

async def init_mqtt(settings: Settings, discoveries) -> Client:
    client = Client()

    if settings.mqtt_username:
        client.username_pw_set(settings.mqtt_username, settings.mqtt_password)
    if log.level == logging.DEBUG:
        client.enable_logger(log)

    connected = threading.Event()
    def on_connect(client: Client, userdata, flags, rc):
        log.info(f"Connected to MQTT server with result code {rc}")
        log.debug("Initializing MQTT discovery")
        for topic, payload in discoveries:
            if settings.reset_discovery:
                log.debug(f"Deleting discovery topic {topic}")
                payload_d = json.loads(payload)
                for comp in payload_d["components"]:
                    del_keys = []
                    for k in payload_d["components"][comp]:
                        if k != "platform":
                            del_keys.append(k)
                    for k in del_keys:
                        del payload_d["components"][comp][k]
                payload_del = json.dumps(payload_d)
                client.publish(topic, payload_del, retain=True)

            log.debug(f"Publishing discovery topic {topic}")
            log.debug(json.dumps(json.loads(payload), indent=2))
            client.publish(topic, payload, retain=True)
        connected.set()

    client.will_set(f"{settings.mqtt_prefix}/status", payload="offline", qos=1, retain=True)
    client.on_connect = on_connect
    client.connect_async(settings.mqtt_host, settings.mqtt_port)
    client.loop_start()
    log.debug("Started MQTT loop")
    connected.wait()
    log.debug("Connected to MQTT server")
    return client

API_CONNECTION_STATUS = r'/api/proxy/1/vehicles/{vin}/connection_status'
API_BODY_CONTROLLER_STATE = r'/api/proxy/1/vehicles/{vin}/body_controller_state'
API_CHARGE_STATE = r'/api/1/vehicles/{vin}/vehicle_data?endpoints=charge_state'
API_CLIMATE_STATE = r'/api/1/vehicles/{vin}/vehicle_data?endpoints=climate_state'
API_COMMAND = r'/api/1/vehicles/{vin}/command/{command}?wait=true'


async def fetch_json(s:requests.Session, base:str, url: str, vin:  str) -> dict:
    resp = s.get(f'{base}{url.format(vin=vin)}')
    if not resp.ok:
        raise ValueError(f"Error fetching {url.format(vin=vin)}: {resp.text}")
    resp_json = resp.json()
    if resp_json['response']['result'] == False:
        raise ValueError(f"Error fetching {url.format(vin=vin)}: {resp_json['response']['reason']}")
    return resp.json()['response']['response']

async def send_command(s:requests.Session, base:str, url: str, vin:  str, command: str, body: str) -> dict:
    resp = s.post(f'{base}{url.format(vin=vin, command=command)}', data=body)
    if not resp.ok:
        raise ValueError(f"Error sending command {command} to {vin}: {resp.text}")
    resp_json = resp.json()
    if resp_json['response']['result'] == False:
        raise ValueError(f"Error sending command {command} to {vin}: {resp_json['response']['reason']}")
    log.debug(resp_json)
    return True

# TODO: Refactor all of this spaghetti code, it's a mess and I hate it... xkcd#1739
async def tesla_ble2mqtt(settings: Settings):
    log.info("Starting Tesla BLE to MQTT proxy")

    discoveries, topic_bindings = init_discovery(settings)
    client = await init_mqtt(settings, discoveries)
    
    # Publish online status
    client.publish(f"{settings.mqtt_prefix}/status", payload="online", qos=1, retain=True)
    start_time = time.time()

    topic_history = { vin: {} for vin in settings.vins }

    s = requests.Session()

    pub_topic = topic_bindings["pub_topic"]
    sub_topic = topic_bindings["sub_topic"]

    commands_lock = threading.Lock()
    commands = []
    new_commands = threading.Event()

    last_error = {}
    async def publish_error(vin: str, error: ValueError):
        if vin in last_error and last_error[vin] == error:
            return
        last_error[vin] = error
        log.error(error.args[0])
        client.publish(f"{settings.mqtt_prefix}/{vin}/last_error/state", payload=error.args[0], qos=1, retain=True)

    def on_message(client: Client, userdata, msg):
        log.debug(f"Received message on topic {msg.topic}: {msg.payload.decode()}")

        if not msg.topic in sub_topic:
            log.error(f"Received message on unknown topic {msg.topic}")
            return
        
        cmd = msg.payload.decode()
        x = cmd

        if not cmd in sub_topic[msg.topic]:
            if '*' in sub_topic[msg.topic]:
                cmd = '*'
            else:
                log.warning(f"Received message with unknown payload {cmd} on topic {msg.topic}")
                return
        
        unformatted_command = sub_topic[msg.topic][cmd].split("|")

        command = unformatted_command[0]
        body = unformatted_command[1] if len(unformatted_command) > 1 else None

        vin = msg.topic.split("/")[-3]
        if body:
            body = body.replace('`vin`', vin)
            body = body.replace('`*`', x)    

        log.debug(f"Received command {cmd} for VIN {vin}: {command} with body {body}")
        with commands_lock:
            if command == 'clear_error':
                if vin in last_error:
                    del last_error[vin]
                client.publish(f"{settings.mqtt_prefix}/{vin}/last_error/state", payload=None, qos=1, retain=True)
                return
            commands.append((vin, command, body))
            new_commands.set()

    client.on_message = on_message

    for topic in sub_topic:
        log.debug(f"Subscribing to topic {topic} with commands {sub_topic[topic]}")
        client.subscribe((topic, 1), qos=1)

    try:
        while True:
            async def process_commands():
                with commands_lock:
                    did_commands = False
                    for vin, command, body in commands:
                        log.info(f"Sending command {command} to VIN {vin}")
                        try:
                            await send_command(s, settings.proxy_host, API_COMMAND, vin, command, body)
                        except ValueError as e:
                            await publish_error(vin, e)
                            continue
                        did_commands = True
                    commands.clear()
                    return did_commands

            async def process_vin(vin: str):
                log.debug(f"Processing VIN {vin}")
                fast_update = False
                fast_update = await process_commands() or fast_update

                try:
                    connection_status = await fetch_json(s, settings.proxy_host, API_CONNECTION_STATUS, vin)
                except ValueError as e:
                    await publish_error(vin, e)
                    return True, False
                
                presence = connection_status is not None and bool(connection_status['address'])
                old_presence = topic_history[vin].get('presence', None)
                if old_presence != presence:
                    log.info(f"Presence for VIN {vin} changed to {presence}")
                    # Presence in handler
                    client.publish(f"{settings.mqtt_prefix}/{vin}/presence/state", payload="ON" if presence else "OFF", qos=1, retain=True)
                    topic_history[vin]['presence'] = presence
                    # Device availability
                    client.publish(f"{settings.mqtt_prefix}/{vin}/status", payload="online" if presence else "offline", qos=1, retain=True)

                fast_update = await process_commands() or fast_update

                if not presence:
                    return False, False
                                
                try:
                    body_controller_state = await fetch_json(s, settings.proxy_host, API_BODY_CONTROLLER_STATE, vin)
                except ValueError as e:
                    await publish_error(vin, e)
                    return True, False
                
                if body_controller_state['vehicle_sleep_status'] != 'VEHICLE_SLEEP_STATUS_AWAKE':
                    log.info(f"Vehicle {vin} is sleeping")
                    # TODO: Handle sleeping vehicle instead of waking it up
                    # with fetching data
                    pass
                
                try:
                    charge_state = await fetch_json(s, settings.proxy_host, API_CHARGE_STATE, vin)
                except ValueError as e:
                    await publish_error(vin, e)
                    return True, False

                
                fast_update = await process_commands() or fast_update

                try:
                    climate_state = await fetch_json(s, settings.proxy_host, API_CLIMATE_STATE, vin)
                except ValueError as e:
                    await publish_error(vin, e)
                    return True, False
                
                fast_update = await process_commands() or fast_update
                
                # Publish state
                pub_state = {
                    "body_controller_state": body_controller_state,
                    "connection_status": connection_status,
                    "vehicle_data": {
                        "charge_state": charge_state["charge_state"],
                        "climate_state": climate_state["climate_state"]
                    }
                }
                for sensor, topic in pub_topic.items():
                    sensor_parts = sensor.split(".")
                    state = pub_state
                    for part in sensor_parts:
                        if part not in state:
                            if sensor != 'uptime':
                                log.error(f"Sensor {sensor} not found in state")
                            state = None
                            break
                        state = state[part]
                    
                    if state is not None:
                        old_state = topic_history[vin].get(sensor, None)
                        if old_state != state:
                            log.info(f"Publishing {sensor} for VIN {vin}: {state}")
                            client.publish(topic, payload=state, qos=1, retain=True)
                            topic_history[vin][sensor] = state

                return fast_update, charge_state['charge_state']['charging_state'] == 'Charging'

            uptime = int(time.time() - start_time)
            client.publish(pub_topic["uptime"], payload=uptime)
            fast_update = False
            car_charging = False
            for vin in settings.vins:
                fast_update, car_charging = await process_vin(vin) or fast_update
            if not fast_update:
                new_commands.wait(settings.poll_interval_charging if car_charging else settings.poll_interval)
                new_commands.clear()
    except asyncio.CancelledError:
        pass
    except KeyboardInterrupt:
        pass

    for vin in settings.vins:
        client.publish(f"{settings.mqtt_prefix}/{vin}/presence/state", payload="OFF", qos=1, retain=True)
        client.publish(f"{settings.mqtt_prefix}/{vin}/status", payload="offline", qos=1, retain=True)

    client.loop_stop()
    log.info("Exiting Tesla BLE to MQTT proxy")
    client.disconnect()



def main(argv):
    settings = parse_args(argv)
    asyncio.run(tesla_ble2mqtt(settings))

if __name__ == "__main__":
    main(sys.argv[1:])