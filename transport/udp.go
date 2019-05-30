package transport

import (
	"context"
	"errors"
	"net"
	"strings"

	"github.com/golang/glog"
	"github.com/mitchellh/mapstructure"
)

type UdpTransport struct {
	Listen        string
	MaxPacketSize int

	ch chan []byte
}

func NewUdpTransport(cfg interface{}) (Transport, error) {
	udp := &UdpTransport{
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

func (r *UdpTransport) Run(ctx context.Context) error {
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

func (r *UdpTransport) serve(ctx context.Context, socket net.PacketConn) {
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

func (r *UdpTransport) Receive() <-chan []byte {
	return r.ch
}

func (r *UdpTransport) Send([]byte) error {
	return nil
}
