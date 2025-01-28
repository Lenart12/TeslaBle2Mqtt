package ha_discovery

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Topic = string
type AccessPath = string
type Command = string
type Action = string

type DevicePublishBindings map[Topic]AccessPath
type DeviceSubscribeBindings map[Topic]map[Command]Action

func configureDevice(x any, formatString func(string) string, pub_topic *DevicePublishBindings, sub_topic *DeviceSubscribeBindings) (any, error) {
	if object_x, ok := x.(map[string]any); ok {
		object := make(map[string]any)
		for key, val := range object_x {
			// Internal sensor configuration
			if strings.HasPrefix(key, "__") {
				// Parse:
				// 	__get_state: access_path -> state_topic: access_path
				// 	__get_state/topic_key: access_path -> {topic_key}_state_topic: access_path
				//  __get_state/!topic_key: access_path -> topic_key: access_path
				if strings.HasPrefix(key, "__get_state") {
					val_s, ok := val.(string)
					if !ok {
						return nil, fmt.Errorf("invalid value for key `%s`", key)
					}
					val_s = formatString(val_s)
					key_parts := strings.SplitN(key, "/", 2)
					state_topic_key := "state_topic"
					topic_is_key := false
					if len(key_parts) > 1 {
						if strings.HasPrefix(key_parts[1], "!") {
							state_topic_key = key_parts[1][1:]
							topic_is_key = true
						} else {
							state_topic_key = key_parts[1] + "_state_topic"
						}
					}
					topic_s := state_topic_key
					if !topic_is_key {
						topic, ok := object_x[state_topic_key]
						if !ok {
							return nil, fmt.Errorf("__get_state using `%s` for `%s` but is not defined", state_topic_key, val_s)
						}
						topic_s, ok = topic.(string)
						if !ok {
							return nil, fmt.Errorf("expected `%s` to be a topic string", state_topic_key)
						}
					}

					topic_f := formatString(topic_s)
					if _, ok := (*pub_topic)[topic_f]; ok {
						return nil, fmt.Errorf("topic `%s` already defined", topic_f)
					}
					(*pub_topic)[topic_f] = val_s
					// Parse:
					// 	__command/command: action -> command_topic: command->action
					// 	__command/topic_key/command: action -> {topic_key}_command_topic: command->action
					//  __command/!topic_key/command: action -> topic_key: command->action
				} else if strings.HasPrefix(key, "__command") {
					val_s, ok := val.(string)
					if !ok {
						return nil, fmt.Errorf("invalid value for key `%s`", key)
					}
					val_s = formatString(val_s)
					key_parts := strings.SplitN(key, "/", 3)
					if len(key_parts) > 3 {
						return nil, fmt.Errorf("invalid key `%s` for __command", key)
					}
					command_topic_key := "command_topic"
					if len(key_parts) > 2 {
						if strings.HasPrefix(key_parts[1], "!") {
							command_topic_key = key_parts[1][1:]
						} else {
							command_topic_key = key_parts[1] + "_command_topic"
						}
					}
					command := key_parts[len(key_parts)-1]
					topic, ok := object_x[command_topic_key]
					if !ok {
						return nil, fmt.Errorf("__command using `%s` for `%s: %s` but is not defined", command_topic_key, key, val_s)
					}
					topic_s, ok := topic.(string)
					if !ok {
						return nil, fmt.Errorf("expected `%s` to be a topic string", command_topic_key)
					}
					topic_f := formatString(topic_s)
					bindings, ok := (*sub_topic)[topic_f]
					if !ok {
						bindings = make(map[Command]Action)
					}
					if _, ok := bindings[command]; ok {
						return nil, fmt.Errorf("command `%s` already defined for topic `%s`", command, topic_f)
					}
					bindings[command] = val_s
					(*sub_topic)[topic_f] = bindings
				}
			} else {
				val_c, err := configureDevice(val, formatString, pub_topic, sub_topic)
				if err != nil {
					return nil, err
				}
				key_f := formatString(key)
				object[key_f] = val_c
			}
		}
		return object, nil
	} else if list_x, ok := x.([]interface{}); ok {
		list := make([]interface{}, len(list_x))
		for i, val_x := range list_x {
			val_c, err := configureDevice(val_x, formatString, pub_topic, sub_topic)
			if err != nil {
				return nil, err
			}
			list[i] = val_c
		}
		return list, nil
	} else if val_s, ok := x.(string); ok {
		return formatString(val_s), nil
	} else {
		return x, nil
	}
}

func GenerateResetConfiguration(original *json.RawMessage) (json.RawMessage, error) {
	var dev map[string]any
	err := json.Unmarshal(*original, &dev)
	if err != nil {
		return nil, err
	}

	components, ok := dev["components"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("components not found in device configuration")
	}

	for key, val := range components {
		// Delete all keys except "platform"
		comp, ok := val.(map[string]any)
		if !ok || comp["platform"] == nil {
			continue
		}
		new_comp := make(map[string]any)
		new_comp["platform"] = comp["platform"]
		components[key] = new_comp
	}

	reset_json, err := json.Marshal(dev)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(reset_json), nil
}

// ParseDeviceConfiguration parses a device configuration and returns the discovery configuration, publish bindings, and subscribe bindings.
// It replaces any string in the configuration with `key` with the values in the replacements map and
// deletes any key that starts with __. It returns bindings for objects with the __get_state or __command key.
func ParseDeviceConfiguration(dev map[string]any, replacements map[string]string) (discovery any, pub_topic DevicePublishBindings, sub_topic DeviceSubscribeBindings, err error) {
	formatString := func(value string) string {
		for key, val := range replacements {
			value = strings.ReplaceAll(value, "`"+key+"`", val)
		}
		return value
	}

	pub_topic = make(DevicePublishBindings)
	sub_topic = make(DeviceSubscribeBindings)

	dev_c, err := configureDevice(dev, formatString, &pub_topic, &sub_topic)
	if err != nil {
		return nil, nil, nil, err
	}

	return dev_c, pub_topic, sub_topic, nil
}
