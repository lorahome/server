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

	ch chan []byte
}

func NewLoRaUdp(cfg interface{}) (LoRaTransport, error) {
	udp := &LoRaUdp{
		ch: make(chan []byte, 1),
	}

	// Map / verify configuration
	err := mapstructure.Decode(cfg, udp)
	if udp.Listen == "" {
		return nil, errors.New("config parameter udp.listen is required")
	}
	if udp.MaxPacketSize == 0 {
		udp.MaxPacketSize = 1024
	}

	return udp, err
}

func (r *LoRaUdp) Run(ctx context.Context) error {
	// Create UDP listening socket
	socket, err := net.ListenPacket("udp", r.Listen)
	if err != nil {
		return err
	}
	glog.Infof("Server started at %s", r.Listen)

	// Receive packets
	go r.serve(ctx, socket)
	// Wait until context canceled
	<-ctx.Done()
	socket.Close()

	return nil
}

func (r *LoRaUdp) serve(ctx context.Context, socket net.PacketConn) {
	buf := make([]byte, r.MaxPacketSize)
	for {
		n, _, err := socket.ReadFrom(buf)
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

func (r *LoRaUdp) Send([]byte) error {
	return nil
}
