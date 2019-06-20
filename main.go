package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/golang/glog"
	"github.com/lorahome/server/db/influxdb"
	"github.com/lorahome/server/devices"
	"github.com/lorahome/server/mqtt"
	"github.com/lorahome/server/transport"

	// Link these devices into server app
	_ "github.com/lorahome/server/devices/sensor/multisensor"
	_ "github.com/lorahome/server/devices/loveheart"
)

var flagConfig = flag.String("config", "config.yaml", "Config filename")
var flagDevices = flag.String("devices", "devices.yaml", "Devices filename")

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()

	// Load configuration
	cfg, err := ConfigLoadFromFile(*flagConfig)
	if err != nil {
		glog.Fatalf("Unable to read config file %s: %v", *flagConfig, err)
	}

	// Setup services
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	caps := &devices.Capabilities{}

	// Start UDP transport
	caps.Udp, err = transport.NewLoRaUdp(cfg.Udp)
	if err != nil {
		glog.Fatalf("LoRa UDP transport failed: %v", err)
	}
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		err := caps.Udp.Run(ctx)
		if err != nil {
			glog.Fatalf("UDP failed: %v", err)
		}
		wg.Done()
	}(&wg)

	// Start InfluxDB
	caps.InfluxDb, err = influxdb.NewInfluxDB(cfg.InfluxDb)
	if err != nil {
		glog.Fatalf("InfluxDB failed: %v", err)
	}

	// Start MQTT client
	caps.Mqtt, err = mqtt.NewMqttClient(cfg.Mqtt)
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		err := caps.Mqtt.Run(ctx)
		if err != nil {
			glog.Fatalf("MQTT failed: %v", err)
		}
		wg.Done()
	}(&wg)

	// Load / register devices
	time.Sleep(100 * time.Millisecond) // Find better solution?
	err = devices.LoadFromFile(*flagDevices, caps)
	if err != nil {
		glog.Fatalf("Unable to load devices: %v", err)
	}

	// Start all devices
	devices.StartAllDevices(ctx)

	// Setup SIGTERM / SIGINT
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		// Wait for packet from any transport
		var err error
		select {
		case packet := <-caps.Udp.Receive():
			err = processPacket(packet)
		case sig := <-signalCh:
			glog.Infof("Got SIG %v", sig)
			// Cancel context and wait until all jobs done
			cancel()
			wg.Wait()
			// Save devices configuration
			err := devices.SaveToFile(*flagDevices)
			if err != nil {
				glog.Fatalf("Save devices config failed: %v", err)
			}
			glog.Info("Gracefully terminated")
			glog.Flush()
			return
		}
		// Handle all errors in one place
		if err != nil {
			glog.Infof("ProcessPacket failed: %v", err)
		}
	}

}
