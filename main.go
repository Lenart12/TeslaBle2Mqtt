package main

import (
	"TeslaBle2Mqtt/internal/discovery"
	"TeslaBle2Mqtt/internal/handler"
	"TeslaBle2Mqtt/internal/settings"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Wrapper for the mqtt logger to use the charmbracelet logger
type MqttLogger struct {
	PrintlnImpl func(v ...interface{})
	PrintfImpl  func(format string, v ...interface{})
}

func (l MqttLogger) Println(v ...interface{}) {
	l.PrintlnImpl(v...)
}
func (l MqttLogger) Printf(format string, v ...interface{}) {
	l.PrintfImpl(format, v...)
}
func NewMqttLogger(level log.Level) MqttLogger {
	return MqttLogger{
		PrintlnImpl: func(v ...interface{}) {
			v_str := make([]string, len(v))
			for i, vv := range v {
				v_str[i] = fmt.Sprintf("%v", vv)
			}
			log.Logf(level, "[MQTT] %s", strings.Join(v_str, " "))
		},
		PrintfImpl: func(format string, v ...interface{}) {
			log.Logf(level, fmt.Sprintf("[MQTT] %s", format), v...)
		},
	}
}

func main() {
	set := settings.Get()
	level, _ := log.ParseLevel(set.LogLevel)
	log.SetLevel(level)
	if set.MqttDebug {
		mqtt.ERROR = NewMqttLogger(log.ErrorLevel)
		mqtt.CRITICAL = NewMqttLogger(log.ErrorLevel)
		mqtt.WARN = NewMqttLogger(log.WarnLevel)
		mqtt.DEBUG = NewMqttLogger(log.DebugLevel)
	}

	log.Info("Starting TeslaBle2Mqtt")

	log.Debug("Running with", "settings", set)
	mqtt.DEBUG.Println("Mqtt debug enabled")

	discoveries, err := discovery.GetDiscovery(set.SensorsYaml, discovery.DiscoverySettings{
		DiscoveryPrefix:  set.DiscoveryPrefix,
		MqttPrefix:       set.MqttPrefix,
		Vins:             set.Vins,
		Version:          "0.0.1",
		ConfigurationUrl: set.ProxyHost + "/dashboard",
	})
	if err != nil {
		log.Fatal("Failed to get discovery", "error", err)
	}
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, d := range discoveries {
		wg.Add(1)
		go handler.Run(ctx, &wg, &d)
	}

	// Wait for all handlers to finish
	success := make(chan bool, 1)
	go func() {
		wg.Wait()
		success <- true
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	select {
	case <-success:
	case <-c:
		log.Info("Received interrupt, shutting down")
		cancel()
		select {
		case <-c:
			log.Fatal("Forced shutdown")
		case <-success:
		}
	}
}
