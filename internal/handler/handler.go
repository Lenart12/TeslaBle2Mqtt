package handler

import (
	"TeslaBle2Mqtt/internal/discovery"
	"TeslaBle2Mqtt/internal/settings"
	"TeslaBle2Mqtt/pkg/ha_discovery"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func publishDiscovery(client mqtt.Client, discovery *discovery.DeviceDiscovery) error {
	s := settings.Get()

	if s.ResetDiscovery {
		reset_discovery, err := ha_discovery.GenerateResetConfiguration(&discovery.Message)
		log.Debug("Resetting discovery", "topic", discovery.Topic, "len", len(reset_discovery))
		if err != nil {
			log.Error("Failed to generate reset discovery", "error", err)
			return err
		}
		json_bytes, err := reset_discovery.MarshalJSON()
		if err != nil {
			log.Error("Failed to marshal reset discovery", "error", err)
			return err
		}
		token := client.Publish(discovery.Topic, s.MqttQos, false, json_bytes)
		token.Wait()
		if token.Error() != nil {
			log.Error("Failed to reset discovery", "error", token.Error())
			return token.Error()
		}
	}

	log.Debug("Publishing discovery", "topic", discovery.Topic, "len", len(discovery.Message))
	json_bytes, err := discovery.Message.MarshalJSON()
	if err != nil {
		log.Error("Failed to marshal discovery", "error", err)
	}
	token := client.Publish(discovery.Topic, s.MqttQos, true, json_bytes)
	token.Wait()
	if token.Error() != nil {
		log.Error("Failed to publish discovery", "error", token.Error())
		return token.Error()
	}
	return nil
}

func publishError(client mqtt.Client, vin string, err error) {
	s := settings.Get()
	error_topic := fmt.Sprintf("%s/%s/last_error/state", s.MqttPrefix, vin)
	error_str := ""
	if err != nil {
		error_str = err.Error()
		// Max length of 255
		if len(error_str) > 255 {
			error_str = error_str[:251]
			error_str += "..."
		}
	} else {
		error_str = "null"
	}
	token := client.Publish(error_topic, s.MqttQos, true, error_str)
	token.Wait()
}

func getProxyResponse(ctx context.Context, http_client *http.Client, method string, endpoint string, body string) (map[string]interface{}, error) {
	s := settings.Get()
	proxy_url := fmt.Sprintf("%s%s", s.ProxyHost, endpoint)
	log.Debug("Getting proxy response", "url", proxy_url)
	var reader io.Reader = nil
	if body != "" {
		reader = strings.NewReader(body)
	}
	request, err := http.NewRequestWithContext(ctx, method, proxy_url, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := http_client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	// log.Debug("Got response", "result", result)

	response, ok := result["response"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no response in result")
	}

	response_result, ok := response["result"].(bool)
	if !ok {
		return nil, fmt.Errorf("no result in response")
	}

	if !response_result {
		return nil, fmt.Errorf("command failed: %s", response["reason"].(string))
	}

	if response_response, ok := response["response"].(map[string]interface{}); ok {
		return response_response, nil
	}

	return nil, nil
}

func handleCommand(ctx context.Context, vin string, http_client *http.Client, mqtt_client mqtt.Client, handler map[ha_discovery.Command]discovery.SubCommand, payload []byte) error {
	command_key := string(payload)

	var command discovery.SubCommand
	var ok bool
	if command, ok = handler[command_key]; !ok {
		def_handler, def := handler["*"]
		if !def {
			return fmt.Errorf("no handler for command `%s`", command_key)
		}
		command_key = "*"
		command = def_handler
	}

	action := command.Command
	body := command.Body
	if body != "" {
		body = strings.ReplaceAll(body, "`*`", string(payload))
	}
	log.Info("Handling command", "key", command_key, "action", action, "body", body)

	if action == "clear_error" {
		publishError(mqtt_client, vin, nil)
		return nil
	}

	endpoint := ""
	// Special case for wake_up
	if action == "wake_up" {
		endpoint = fmt.Sprintf("/api/1/vehicles/%s/wake_up?wait=true", vin)
	} else {
		endpoint = fmt.Sprintf("/api/1/vehicles/%s/command/%s?wait=true", vin, action)
	}
	_, err := getProxyResponse(ctx, http_client, http.MethodPost, endpoint, body)
	if err != nil {
		return err
	}
	log.Debug("Command handled successfuly", "key", command_key, "action", action, "body", body)

	return nil
}

var uptime_start *time.Time

func getState(ctx context.Context, vin string, http_client *http.Client, device_type discovery.DeviceType, pub *discovery.DevicePublishBindings) (map[string]string, error) {
	state := make(map[string]any)
	topicState := make(map[string]string)
	if device_type == discovery.HandlerDeviceType {
		if uptime_start == nil {
			t0 := time.Now()
			uptime_start = &t0
		}
		uptime := time.Since(*uptime_start)
		state["status"] = "online" // Always online
		state["uptime"] = fmt.Sprintf("%d", int(uptime.Seconds()))
	} else if device_type == discovery.PerVehicleDeviceType {
		state["status"] = "offline"

		// Get connection status
		connection_status_url := fmt.Sprintf("/api/proxy/1/vehicles/%s/connection_status", vin)
		connection_status, err := getProxyResponse(ctx, http_client, http.MethodGet, connection_status_url, "")
		if err != nil {
			return nil, fmt.Errorf("failed to get connection status: %w", err)
		}
		state["connection_status"] = connection_status
		// If the vehicle is in range, get body controller state
		if connection_status["address"] != nil {
			state["status"] = "online"
			body_controller_state_url := fmt.Sprintf("/api/proxy/1/vehicles/%s/body_controller_state", vin)
			body_controller_state, err := getProxyResponse(ctx, http_client, http.MethodGet, body_controller_state_url, "")
			if err != nil {
				return nil, fmt.Errorf("failed to get body controller state: %w", err)
			}
			state["body_controller_state"] = body_controller_state
			// If the vehicle is awake, get vehicle state
			if body_controller_state["vehicle_sleep_status"] == "VEHICLE_SLEEP_STATUS_AWAKE" {
				vehicle_state_url := fmt.Sprintf("/api/1/vehicles/%s/vehicle_data?endpoints=charge_state;climate_state", vin)
				vehicle_state, err := getProxyResponse(ctx, http_client, http.MethodGet, vehicle_state_url, "")
				if err != nil {
					return nil, fmt.Errorf("failed to get vehicle state: %w", err)
				}
				state["vehicle_data"] = vehicle_state

				if cs, ok := vehicle_state["charge_state"].(map[string]any); ok {
					if cs["charging_state"] == "Charging" {
						topicState["poll_interval"] = strconv.Itoa(settings.Get().PollIntervalCharging) // Special case for charging poll interval
						log.Debug("Using charging poll interval", "vin", vin)
					}
				}
			} else {
				log.Debug("Vehicle not awake", "vin", vin)
			}
		} else {
			log.Debug("Vehicle not in range", "vin", vin)
		}
	} else {
		log.Error("Invalid device type", "device_type", device_type)
		return nil, fmt.Errorf("invalid device type")
	}
	for topic, access_path := range *pub {
		value := "None"

		access_path_parts := strings.Split(access_path, ".")
		current_state := state
		for _, part := range access_path_parts {
			// Traverse the state object
			if current_state == nil {
				break
			}
			if next, ok := current_state[part]; ok {
				if next_map, ok := next.(map[string]any); ok {
					current_state = next_map
				} else {
					value = fmt.Sprintf("%v", next)
					if value == "<nil>" {
						value = "null"
					}
					break
				}
			} else {
				break
			}
		}
		// log.Debug("Processed new", "topic", topic, "access_path", access_path, "value", value)
		topicState[topic] = value
	}
	topicState["status"] = state["status"].(string) // Special case for status
	return topicState, nil
}

func publishState(ctx context.Context, vin string, http_client *http.Client, mqtt_client mqtt.Client, disc *discovery.DiscoveryHandler, old_state map[ha_discovery.Topic]string, onlineHysteresis *int) (time.Duration, error) {
	start := time.Now()
	s := settings.Get()

	poll_interval := time.Duration(s.PollInterval) * time.Second
	skip_publish := false

	log.Debug("Getting state", "handler", disc.ClientId)
	state, err := getState(ctx, vin, http_client, disc.Discovery.DeviceType, &disc.PublishBindings)
	// log.Debug("Got state", "state", state)

	if err != nil {
		poll_interval = time.Duration(1) * time.Second

		// Happens when vehicle was in range for connection_status, but not for body_controller_state and vehicle_data
		if strings.Contains(err.Error(), "vehicle not in range") {
			state = make(map[string]string)
			state["status"] = "offline"
			err = nil
		} else {
			if ctx.Err() != nil {
				log.Warn("Timed out getting state", "handler", disc.ClientId)
				return poll_interval, ctx.Err()
			}
			log.Warn("Failed to get state", "handler", disc.ClientId, "error", err)
			return poll_interval, err
		}
	}

	if state["status"] == "online" {
		*onlineHysteresis = 3
	} else {
		*onlineHysteresis = max(*onlineHysteresis-1, 0)
		// Keep online
		if *onlineHysteresis > 0 {
			log.Debug("Device going offline", "handler", disc.ClientId, "hysteresis", *onlineHysteresis)
			skip_publish = true
			poll_interval = time.Duration(1) * time.Second // Retry quickly when going offline
		}
	}

	if err == nil && !skip_publish {
		for topic, access_path := range disc.PublishBindings {
			// If new state is different from old state, publish
			if new_state, ok := state[topic]; ok {
				if new_state != old_state[topic] {
					log.Info("Publishing", "topic", topic, "access path", access_path, "state", new_state, "old_state", old_state[topic])
					token := mqtt_client.Publish(topic, s.MqttQos, true, new_state)
					select {
					case <-ctx.Done():
						return time.Duration(1) * time.Second, ctx.Err()
					case <-token.Done():
					}
					if token.Error() != nil {
						log.Error("Failed to publish state", "error", token.Error())
						return time.Duration(1) * time.Second, token.Error()
					}
					old_state[topic] = new_state
				}
			}
		}
		if pi, ok := state["poll_interval"]; ok {
			pi_int, err := strconv.Atoi(pi)
			if err == nil {
				poll_interval = time.Duration(pi_int) * time.Second
			} else {
				log.Error("Failed to parse poll_interval", "error", err)
			}
		}
	} else if err != nil {
		log.Debug("Failed to get state", "error", err)
		return poll_interval, err
	}
	return poll_interval - time.Since(start), nil
}

// Run is the main handler function, it takes a discovery object and runs the handler
// for the given device. It will handle the discovery and mqtt communication for the
// given device along with fetching and updating the device state defined in the discovery bindings.
// This function should be run as a goroutine.
func Run(ctx context.Context, wg *sync.WaitGroup, disc *discovery.DiscoveryHandler) {
	log.Debug("Running", "handler", disc.Discovery.DeviceType, "for", disc.ClientId)
	defer wg.Done()

	s := settings.Get()

	clear_old_state_request := false
	clientOpts := mqtt.NewClientOptions().
		AddBroker(fmt.Sprintf("tcp://%s:%d", s.MqttHost, s.MqttPort)).
		SetUsername(s.MqttUser).
		SetPassword(s.MqttPass).
		SetClientID(disc.ClientId).
		SetWill(disc.WillTopic, "offline", s.MqttQos, true).
		SetOnConnectHandler(func(client mqtt.Client) {
			log.Info("Connected to MQTT", "client_id", disc.ClientId)
			clear_old_state_request = true
			if err := publishDiscovery(client, &disc.Discovery); err != nil {
				log.Error("Failed to publish discovery", "error", err)
				return
			}
		}).
		SetConnectionLostHandler(func(client mqtt.Client, err error) {
			log.Error("Connection lost to MQTT", "client_id", disc.ClientId, "error", err)
		})

	mqtt_client := mqtt.NewClient(clientOpts)

	if token := mqtt_client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("Failed to connect to MQTT", "error", token.Error())
		return
	}

	defer mqtt_client.Disconnect(250)
	defer func() {
		token := mqtt_client.Publish(disc.WillTopic, s.MqttQos, true, "offline")
		token.Wait()
	}()

	to_subscribe := map[string]byte{}
	for topic := range disc.SubscribeBindings {
		to_subscribe[topic] = s.MqttQos
	}
	// Subscribe to status topic to resend discovery on HA restart
	ha_status_topic := fmt.Sprintf("%s/status", s.DiscoveryPrefix)
	to_subscribe[ha_status_topic] = s.MqttQos

	message_chan := make(chan mqtt.Message, 10)

	mqtt_client.SubscribeMultiple(to_subscribe, func(client mqtt.Client, msg mqtt.Message) {
		message_chan <- msg
	})

	cancel_get_state_ch := make(chan bool)

	http_client := &http.Client{}

	// Publish loop
	go func() {
		old_state := make(map[ha_discovery.Topic]string)
		onlineHysteresis := 0
	start_publish:
		for {
			publishCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			done := make(chan bool)
			go func() {
				if clear_old_state_request {
					log.Debug("Clearing old state", "handler", disc.ClientId)
					old_state = make(map[ha_discovery.Topic]string)
					clear_old_state_request = false
				}
				to_wait, err := publishState(publishCtx, disc.Vin, http_client, mqtt_client, disc, old_state, &onlineHysteresis)
				if err != nil && err != publishCtx.Err() {
					log.Error("Failed to publish state", "error", err)
					publishError(mqtt_client, disc.Vin, err)
				}
				select {
				case <-time.After(to_wait):
				case <-publishCtx.Done():
				}
				done <- true
			}()
			for {
				select {
				case <-done:
					continue start_publish
				case <-publishCtx.Done():
					// If the context is done, we need to check if the main context is done
					// and if so, exit the loop
					if ctx.Err() != nil {
						return
					}
					continue start_publish
				case end := <-cancel_get_state_ch:
					if end {
						log.Debug("Canceling publish", "handler", disc.ClientId)
						cancel()
						select {
						case <-ctx.Done():
							return
						case <-cancel_get_state_ch:
							log.Debug("Resuming publish", "handler", disc.ClientId)
							continue start_publish
						}
					} else {
						log.Warn("Got signal to get state, but is already doing so", "handler", disc.ClientId)
					}
				// Just in case the publish loop gets stuck, we will timeout after 20+interval seconds and restart
				case <-time.After(time.Duration(20+s.PollInterval) * time.Second):
					log.Warn("Publish loop timed out", "handler", disc.ClientId)
					cancel()
					continue start_publish
				}

			}
		}
	}()

handler_loop:
	for {
		select {
		case <-ctx.Done():
			log.Debug("Context done, shutting down", "handler", disc.ClientId)
			break handler_loop
		case msg := <-message_chan:
			log.Debug("Received message", "topic", msg.Topic(), "message", string(msg.Payload()))

			if msg.Topic() == ha_status_topic {
				log.Debug("HA status changed", "to", string(msg.Payload()))
				if string(msg.Payload()) != "online" {
					continue
				}
				log.Info("Resending discovery", "topic", ha_status_topic)
				if err := publishDiscovery(mqtt_client, &disc.Discovery); err != nil {
					log.Error("Failed to publish discovery", "error", err)
				}
				clear_old_state_request = true
				continue
			}

			if handler, ok := disc.SubscribeBindings[msg.Topic()]; ok {
				cancel_get_state_ch <- true
				err := handleCommand(ctx, disc.Vin, http_client, mqtt_client, handler, msg.Payload())
				cancel_get_state_ch <- false
				if err != nil {
					log.Error("Failed to handle command", "error", err)
					publishError(mqtt_client, disc.Vin, err)
				}
			} else {
				log.Warn("No handler for message", "topic", msg.Topic())
			}
		}
	}
}
