import json
from typing import Dict, List
import yaml
from teslable2mqtt import Settings
import const
import logging

log = logging.getLogger()

def parse_config(config: str) -> dict:
    """
    Parse a configuration file in YAML format
    """
    with open(config, "r") as f:
        return yaml.safe_load(f)
    
def generate_discovery(settings: Settings, config: dict) -> Dict[str, str]:
    """
    Generate Home Assistant MQTT Discovery configuration
    """

    def format_str(value: str, vin: str|None) -> str:
        replacements = {
            "mqtt_prefix": settings.mqtt_prefix,
            "vin": vin if vin else '`vin`',
            "tb2m_version": const.VERSION,
            "tb2m_configuration_url": f'{settings.proxy_host}/dashboard',
            "vehicle_model": vin[3] if vin else '`vehicle_model`',
        }
        for key, val in replacements.items():
            value = value.replace(f"`{key}`", val)
        return value

    pub_topic = {}
    sub_topic = {}

    def configure_device(device: dict, vin: str|None) -> dict:
        device = dict(device)
        del_keys = []
        for key, val in device.items():
            if key.startswith("__"):
                if key == "__get_state":
                    pub_topic[val] = format_str(device["state_topic"], vin)
                elif key.startswith("__command"):
                    cmd = key.split('/')[-1]
                    topic = format_str(device["command_topic"], vin)
                    if topic not in sub_topic:
                        sub_topic[topic] = {}
                    sub_topic[topic][cmd] = format_str(val, vin)
                del_keys.append(key)
                continue
            elif isinstance(val, str):
                device[key] = format_str(val, vin)
            elif isinstance(val, list):
                device[key] = [format_str(v, vin) if isinstance(v, str) else configure_device(v, vin)  for v in val]
            elif isinstance(val, dict):
                device[key] = configure_device(val, vin)
        for key in del_keys:
            del device[key]
        return device
    
    
    discovery_devices = {}
    handler = configure_device(config["devices"]["handler"], None)

    for vin in settings.vins:
        vehicle_device = configure_device(config["devices"]["per_vehicle"], vin)
        handler["components"][f"presence_{vin}"] = configure_device(config["devices"]["handler"]["components"]["__presence_template"], vin)
        discovery_devices[f'{settings.discovery_prefix}/device/{settings.mqtt_prefix}_{vin}/config'] = json.dumps(vehicle_device)
        log.debug(f"Generated device for VIN {vin}")

    discovery_devices[f'{settings.discovery_prefix}/device/tb2m/config'] = json.dumps(handler)

    return discovery_devices, {
        "pub_topic": pub_topic,
        "sub_topic": sub_topic
    }