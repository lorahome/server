package transport

import (
	"context"
)

type Transport interface {
	Run(context.Context) error
	Receive() <-chan []byte
	Send([]byte) error
}
