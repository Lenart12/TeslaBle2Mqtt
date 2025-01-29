package discovery

import (
	"TeslaBle2Mqtt/pkg/ha_discovery"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"gopkg.in/yaml.v3"
)

//go:embed mqtt_sensors.yaml
var default_config []byte

func loadYamlFile(filename string) (map[string]interface{}, error) {
	var file_data []byte
	if filename == "" {
		file_data = default_config
		log.Debug("Using default sensors configuration")
	} else {
		var err error
		file_data, err = os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		log.Debug("Using sensors configuration from", "filename", filename)
	}

	sensors_config := make(map[string]interface{})
	err := yaml.Unmarshal(file_data, &sensors_config)

	if err != nil {
		return nil, err
	}
	return sensors_config, nil
}

type SubCommand struct {
	Command string
	Body    string
}

func parseSubCommand(command string) (SubCommand, error) {
	var cmd SubCommand
	command_parts := strings.Split(command, "|")
	if len(command_parts) > 2 {
		return cmd, fmt.Errorf("invalid command format - `|` (%s)", command)
	}
	cmd.Command = command_parts[0]
	if len(command_parts) == 2 {
		cmd.Body = command_parts[1]
	}
	return cmd, nil
}

type DeviceType string

const (
	HandlerDeviceType    DeviceType = "handler"
	PerVehicleDeviceType DeviceType = "per_vehicle"
)

type DevicePublishBindings = ha_discovery.DevicePublishBindings
type DeviceSubscribeBindings = map[ha_discovery.Topic]map[ha_discovery.Command]SubCommand

type DeviceDiscovery struct {
	Topic      ha_discovery.Topic
	DeviceType DeviceType
	Message    json.RawMessage
}

type DiscoveryHandler struct {
	Discovery         DeviceDiscovery
	Vin               string
	ClientId          string
	WillTopic         string
	PublishBindings   DevicePublishBindings
	SubscribeBindings DeviceSubscribeBindings
}

type DiscoverySettings struct {
	DiscoveryPrefix  string
	MqttPrefix       string
	Vins             []string
	Version          string
	ConfigurationUrl string
}

func vehicleModel(vin byte) string {
	switch vin {
	case 'S':
		return "Model S"
	case '3':
		return "Model 3"
	case 'X':
		return "Model X"
	case 'Y':
		return "Model Y"
	case 'C':
		return "Cybertruck"
	case 'R':
		return "Roadster"
	case 'T':
		return "Semi"
	default:
		log.Fatal("Unknown vehicle model (or invalid VIN)")
		return ""
	}
}

func GetDiscovery(filename string, settings DiscoverySettings) ([]DiscoveryHandler, error) {
	sensors_config, err := loadYamlFile(filename)
	if err != nil {
		return nil, err
	}

	devices, ok := sensors_config["devices"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("devices not found or invalid in %s", filename)
	}

	discoveries := make([]DiscoveryHandler, 0)

	discoveryTopic := func(device_id string) string {
		return fmt.Sprintf("%s/device/%s/config", settings.DiscoveryPrefix, device_id)
	}

	parseDeviceWithPubSub := func(device *map[string]interface{}, replacements map[string]string) (any, DevicePublishBindings, DeviceSubscribeBindings, error) {
		disc, pub, sub, err := ha_discovery.ParseDeviceConfiguration(*device, replacements)
		if err != nil {
			return nil, nil, nil, err
		}

		sub_cmd := DeviceSubscribeBindings{}
		for topic, commands := range sub {
			sub_cmd[topic] = make(map[ha_discovery.Command]SubCommand)
			for command_key, action := range commands {
				sub_command, err := parseSubCommand(action)
				if err != nil {
					return nil, nil, nil, err
				}
				sub_cmd[topic][command_key] = sub_command
			}
		}
		return disc, pub, sub_cmd, nil
	}

	addDevice := func(device *map[string]interface{}, vin string, clientId string, discovery_topic string, will_topic string, device_type DeviceType, replacements map[string]string) error {
		disc, pub, sub, err := parseDeviceWithPubSub(device, replacements)
		if err != nil {
			return err
		}
		disc_json, err := json.Marshal(disc)
		if err != nil {
			return err
		}
		discoveries = append(discoveries, DiscoveryHandler{
			Discovery: DeviceDiscovery{
				Topic:      discovery_topic,
				DeviceType: device_type,
				Message:    disc_json,
			},
			Vin:               vin,
			ClientId:          clientId,
			WillTopic:         will_topic,
			PublishBindings:   pub,
			SubscribeBindings: sub,
		})
		return nil
	}

	vinReplacements := func(vin string) map[string]string {
		return map[string]string{
			"vin":                    vin,
			"mqtt_prefix":            settings.MqttPrefix,
			"tb2m_version":           settings.Version,
			"tb2m_configuration_url": settings.ConfigurationUrl,
			"vehicle_model":          vehicleModel(vin[3]),
		}
	}

	handler, ok := devices["handler"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("handler not found or invalid in %s", filename)
	}
	handler_vin_components, ok := devices["handler_vin_components"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("handler_vin_components not found or invalid in %s", filename)
	}
	for _, vin := range settings.Vins {
		comp_disc, _, _, err := parseDeviceWithPubSub(&handler_vin_components, vinReplacements(vin))
		// TODO: Do not ignore publish and subscribe bindings
		if err != nil {
			return nil, err
		}
		comps, ok := handler["components"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("components not found or invalid in %s", filename)
		}
		handler_comps, ok := comp_disc.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("components not found or invalid in %s", filename)
		}
		for comp_id, comp := range handler_comps {
			comps[comp_id] = comp
		}
	}

	if err := addDevice(&handler, "tb2m", settings.MqttPrefix, discoveryTopic(settings.MqttPrefix),
		fmt.Sprintf("%s/status", settings.MqttPrefix),
		HandlerDeviceType, map[string]string{
			"mqtt_prefix":            settings.MqttPrefix,
			"tb2m_version":           settings.Version,
			"tb2m_configuration_url": settings.ConfigurationUrl,
		}); err != nil {
		return nil, err
	}

	per_vehicle, ok := devices["per_vehicle"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("per_vehicle not found or invalid in %s", filename)
	}

	for _, vin := range settings.Vins {
		clientId := fmt.Sprintf("%s_%s", settings.MqttPrefix, vin)
		if err := addDevice(&per_vehicle, vin, clientId, discoveryTopic(clientId),
			fmt.Sprintf("%s/%s/status", settings.MqttPrefix, vin), PerVehicleDeviceType,
			vinReplacements(vin)); err != nil {
			return nil, err
		}
	}

	return discoveries, nil
}
