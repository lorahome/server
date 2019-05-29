package transport

import (
	"context"
	"net"
	"strings"

	"github.com/golang/glog"
	"github.com/lorahome/server/config"
)

type UdpTransport struct {
	config *config.Udp
	ch     chan []byte
}

func NewUdpTransport(config *config.Config) Transport {
	return &UdpTransport{
		config: &config.Udp,
		ch:     make(chan []byte, 1),
	}
}

func (r *UdpTransport) Run(ctx context.Context) error {
	// Create UDP listening socket
	socket, err := net.ListenPacket("udp", r.config.Listen)
	if err != nil {
		return err
	}
	glog.Infof("[udp] server started at %s", r.config.Listen)

	// Receive packets
	go r.serve(ctx, socket)
	// Wait until context canceled
	<-ctx.Done()
	socket.Close()

	return nil
}

func (r *UdpTransport) serve(ctx context.Context, socket net.PacketConn) {
	buf := make([]byte, r.config.MaxPacketSize)
	for {
		n, _, err := socket.ReadFrom(buf)
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			glog.Infof("[udp] readFrom failed: %v", err)
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
