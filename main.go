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

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/muesli/termenv"
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

func ansi16Style() *log.Styles {
	return &log.Styles{
		Timestamp: lipgloss.NewStyle(),
		Caller:    lipgloss.NewStyle().Foreground(lipgloss.Color("0")),
		Prefix:    lipgloss.NewStyle().Foreground(lipgloss.Color("0")),
		Message:   lipgloss.NewStyle(),
		Key:       lipgloss.NewStyle().Foreground(lipgloss.Color("0")),
		Value:     lipgloss.NewStyle(),
		Separator: lipgloss.NewStyle().Foreground(lipgloss.Color("0")),
		Levels: map[log.Level]lipgloss.Style{
			log.DebugLevel: lipgloss.NewStyle().
				SetString(strings.ToUpper(log.DebugLevel.String())).
				MaxWidth(4).
				Foreground(lipgloss.Color("5")),
			log.InfoLevel: lipgloss.NewStyle().
				SetString(strings.ToUpper(log.InfoLevel.String())).
				MaxWidth(4).
				Foreground(lipgloss.Color("6")),
			log.WarnLevel: lipgloss.NewStyle().
				SetString(strings.ToUpper(log.WarnLevel.String())).
				MaxWidth(4).
				Foreground(lipgloss.Color("3")),
			log.ErrorLevel: lipgloss.NewStyle().
				SetString(strings.ToUpper(log.ErrorLevel.String())).
				MaxWidth(4).
				Foreground(lipgloss.Color("1")),
			log.FatalLevel: lipgloss.NewStyle().
				SetString(strings.ToUpper(log.FatalLevel.String())).
				MaxWidth(4).
				Foreground(lipgloss.Color("1")),
		},
		Keys:   map[string]lipgloss.Style{},
		Values: map[string]lipgloss.Style{},
	}
}

func main() {
	set := settings.Get()
	// Set up logging
	if set.ForceAnsiColor {
		log.Default().SetColorProfile(termenv.ANSI)
		log.SetStyles(ansi16Style())
	}
	level, _ := log.ParseLevel(set.LogLevel)
	log.SetLevel(level)
	if set.MqttDebug {
		mqtt.ERROR = NewMqttLogger(log.ErrorLevel)
		mqtt.CRITICAL = NewMqttLogger(log.ErrorLevel)
		mqtt.WARN = NewMqttLogger(log.WarnLevel)
		mqtt.DEBUG = NewMqttLogger(log.DebugLevel)
	}
	log.SetPrefix(set.LogPrefix)

	log.Info("Starting TeslaBle2Mqtt")

	log.Debug("Running with", "settings", set)
	mqtt.DEBUG.Println("Mqtt debug enabled")

	configUrl := set.ReportedConfigUrl
	if configUrl == "{proxy-host}/dashboard" {
		configUrl = set.ProxyHost + "/dashboard"
	}

	discoveries, err := discovery.GetDiscovery(set.SensorsYaml, discovery.DiscoverySettings{
		DiscoveryPrefix:  set.DiscoveryPrefix,
		MqttPrefix:       set.MqttPrefix,
		Vins:             set.Vins,
		Version:          set.ReportedVersion,
		ConfigurationUrl: configUrl,
		MaxChargingAmps:  fmt.Sprintf("%d", set.MaxChargingAmps),
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
