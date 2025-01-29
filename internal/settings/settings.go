package settings

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/akamensky/argparse"
	"github.com/charmbracelet/log"
)

type Settings struct {
	Vins                 []string
	ProxyHost            string
	PollInterval         int
	PollIntervalCharging int
	MqttHost             string
	MqttPort             int
	MqttUser             string
	MqttPass             string
	MqttQos              byte
	DiscoveryPrefix      string
	MqttPrefix           string
	ResetDiscovery       bool
	SensorsYaml          string
	LogLevel             string
	MqttDebug            bool
}

var settings *Settings

func Get() *Settings {
	if settings == nil {
		settings = &Settings{}
		parseSettings(settings)
	}
	return settings
}

func parseSettings(settings *Settings) {
	parser := argparse.NewParser("Tesla BLE to Mqtt", "Expose Tesla sensors and controls to MQTT with Home Assistant discovery")
	vins := parser.List("v", "vin", &argparse.Options{Required: true, Help: "VIN of the Tesla vehicle (Can be specified multiple times)", Validate: func(args []string) error {
		for _, vin := range args {
			if len(vin) != 17 {
				return fmt.Errorf("invalid VIN (%s)", vin)
			}
		}
		return nil
	}})
	proxy_host := parser.String("p", "proxy-host", &argparse.Options{Required: false, Help: "Proxy host", Default: "http://localhost:8080", Validate: func(args []string) error {
		// Check if the proxy host is a valid URL
		url, err := url.Parse(args[0])
		if err != nil {
			return fmt.Errorf("invalid proxy host (%s)", err)
		}
		if url.Scheme != "http" && url.Scheme != "https" {
			return fmt.Errorf("invalid proxy host scheme")
		}
		if url.Path != "" {
			return fmt.Errorf("invalid proxy host path")
		}
		return nil
	}})
	poll_interval := parser.Int("i", "poll-interval", &argparse.Options{Required: false, Help: "Poll interval in seconds", Default: 90})
	poll_interval_charging := parser.Int("I", "poll-interval-charging", &argparse.Options{Required: false, Help: "Poll interval in seconds when charging", Default: 20})
	mqtt_host := parser.String("H", "mqtt-host", &argparse.Options{Required: false, Help: "MQTT host", Default: "localhost"})
	mqtt_port := parser.Int("P", "mqtt-port", &argparse.Options{Required: false, Help: "MQTT port", Default: 1883})
	mqtt_user := parser.String("u", "mqtt-user", &argparse.Options{Required: false, Help: "MQTT username"})
	mqtt_pass := parser.String("w", "mqtt-pass", &argparse.Options{Required: false, Help: "MQTT password"})
	mqtt_qos := parser.Int("q", "mqtt-qos", &argparse.Options{Required: false, Help: "MQTT QoS", Default: 0, Validate: func(args []string) error {
		qos, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid QoS")
		}
		if qos < 0 || qos > 2 {
			return fmt.Errorf("invalid QoS")
		}
		return nil
	}})
	discovery_prefix := parser.String("d", "discovery-prefix", &argparse.Options{Required: false, Help: "MQTT discovery prefix", Default: "homeassistant"})
	mqtt_prefix := parser.String("m", "mqtt-prefix", &argparse.Options{Required: false, Help: "MQTT prefix", Default: "tb2m"})
	sensors_yaml := parser.String("y", "sensors-yaml", &argparse.Options{Required: false, Help: "Path to custom sensors YAML file", Default: ""})
	reset_discovery := parser.Flag("r", "reset-discovery", &argparse.Options{Required: false, Help: "Reset MQTT discovery"})
	log_level := parser.String("l", "log-level", &argparse.Options{Required: false, Help: "Log level", Default: "INFO", Validate: func(args []string) error {
		if _, err := log.ParseLevel(args[0]); err != nil {
			return err
		}
		return nil
	}})
	mqtt_debug := parser.Flag("D", "mqtt-debug", &argparse.Options{Required: false, Help: "Enable MQTT debug output (sam log level as --log-level)"})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	settings.LogLevel = *log_level
	settings.Vins = *vins
	settings.ProxyHost = *proxy_host
	settings.PollInterval = *poll_interval
	settings.PollIntervalCharging = *poll_interval_charging
	settings.MqttHost = *mqtt_host
	settings.MqttPort = *mqtt_port
	settings.MqttUser = *mqtt_user
	settings.MqttPass = *mqtt_pass
	settings.MqttQos = byte(*mqtt_qos)
	settings.DiscoveryPrefix = *discovery_prefix
	settings.MqttPrefix = *mqtt_prefix
	settings.ResetDiscovery = *reset_discovery
	settings.SensorsYaml = *sensors_yaml
	settings.MqttDebug = *mqtt_debug
}
