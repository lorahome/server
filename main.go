package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/golang/glog"
	"github.com/lorahome/server/devices"
	"github.com/lorahome/server/transport"

	// Link these devices into server app
	_ "github.com/lorahome/server/devices/sensor/multisensor"
)

var flagConfig = flag.String("config", "config.yaml", "Config filename")
var flagDevices = flag.String("devices", "devices.yaml", "Devices filename")

type RuntimeCapabilities struct {
	udp transport.LoRaTransport
}

func (rc *RuntimeCapabilities) SendPacket(packet []byte) error {
	// TODO: actual send
	return nil
}

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()

	// Load configuration
	cfg, err := ConfigLoadFromFile(*flagConfig)
	if err != nil {
		glog.Fatalf("Unable to read config file %s: %v", *flagConfig, err)
	}

	// Load / register devices
	err = devices.LoadFromFile(*flagDevices)
	if err != nil {
		glog.Fatalf("Unable to load devices: %v", err)
	}

	// Setup capabilities for devices
	caps := &RuntimeCapabilities{}
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// Start UDP transport
	caps.udp, err = transport.NewLoRaUdp(cfg.Udp)
	if err != nil {
		glog.Fatalf("LoRa UDP transport failed: %v", err)
	}
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		err := caps.udp.Run(ctx)
		if err != nil {
			glog.Fatalf("UDP failed: %v", err)
		}
		glog.Info("UDP terminated")
		wg.Done()
	}(&wg)

	// Setup SIGTERM / SIGINT
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		// Wait for packet from any transport
		var err error
		select {
		case packet := <-caps.udp.Receive():
			err = processPacket(caps, packet)
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
