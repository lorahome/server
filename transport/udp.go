package transport

import (
	"context"
	"errors"
	"net"
	"strings"

	"github.com/golang/glog"
	"github.com/mitchellh/mapstructure"
)

type LoRaUdp struct {
	Listen        string
	MaxPacketSize int
	Gateway       string

	ch                     chan []byte
	enabled                bool
	resolvedGatewayAddress *net.UDPAddr
	socket                 net.PacketConn
}

func NewLoRaUdp(cfg interface{}) (LoRaTransport, error) {
	udp := &LoRaUdp{
		ch: make(chan []byte, 1),
	}

	// If no configuration present - bypass mode
	if cfg == nil {
		return udp, nil
	}

	// Map / verify configuration
	err := mapstructure.Decode(cfg, udp)
	if udp.Listen == "" {
		return nil, errors.New("config parameter udp.listen is required")
	}
	if udp.MaxPacketSize == 0 {
		udp.MaxPacketSize = 1024
	}
	udp.enabled = true

	// Resolve LoRa gateway address, if any
	if udp.Gateway != "" {
		udp.resolvedGatewayAddress, err = net.ResolveUDPAddr("udp", udp.Gateway)
		if err != nil {
			return nil, err
		}
	}

	return udp, err
}

func (r *LoRaUdp) Run(ctx context.Context) error {
	// Simple blocking call for bypass mode
	if !r.enabled {
		glog.Info("UDP is not enabled")
		<-ctx.Done()
		return nil
	}

	// Create UDP listening socket
	var err error
	r.socket, err = net.ListenPacket("udp", r.Listen)
	if err != nil {
		return err
	}
	glog.Infof("UDP server started at %s", r.Listen)
	// Start receiver
	go r.serve(ctx)

	// Wait until context canceled
	<-ctx.Done()
	r.socket.Close()

	return nil
}

func (r *LoRaUdp) serve(ctx context.Context) {
	buf := make([]byte, r.MaxPacketSize)
	for {
		n, _, err := r.socket.ReadFrom(buf)
		if err != nil {
			// Terminate goroutine when listener closed
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			glog.Infof("readFrom failed: %v", err)
			continue
		}
		r.ch <- buf[:n]
	}
}

func (r *LoRaUdp) Receive() <-chan []byte {
	return r.ch
}

func (r *LoRaUdp) Send(packet []byte) error {
	// Just do nothing in bypass mode
	if !r.enabled || r.resolvedGatewayAddress == nil {
		return nil
	}

	_, err := r.socket.WriteTo(packet, r.resolvedGatewayAddress)
	return err
}
