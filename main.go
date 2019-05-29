package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/golang/glog"
	"github.com/lorahome/server/config"
	"github.com/lorahome/server/registry"
	"github.com/lorahome/server/transport"
)

var flagConfig = flag.String("config", "config.yaml", "Config filename")

type RuntimeCapabilities struct {
	udp transport.Transport
}

func (rc *RuntimeCapabilities) SendPacket(packet []byte) error {
	// TODO: actual send
	return nil
}

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfigFromFile(*flagConfig)
	if err != nil {
		glog.Fatalf("Unable to read config file %s: %v", *flagConfig, err)
	}
	// Register known devices from config (already been joined)
	for _, d := range cfg.Devices {
		_, err := registry.RegisterDevice(d)
		if err != nil {
			glog.Fatalf("Unable to register device: %v", err)
		}
	}

	// Setup capabilities for devices
	caps := &RuntimeCapabilities{}
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// Start UDP transport
	caps.udp = transport.NewUdpTransport(cfg)
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
			cancel()
			wg.Wait()
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

// cfg.Devices = registry.GetDevicesForConfigSave()
// config.SaveToFile("zz.yaml", cfg)
