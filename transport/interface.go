package transport

import (
	"context"
)

type LoRaTransport interface {
	Run(context.Context) error
	Receive() <-chan []byte
	Send([]byte) error
}
