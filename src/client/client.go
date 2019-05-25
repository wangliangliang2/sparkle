package client

import (
	"av"
)

type Client interface {
	Pull() (p av.Packet, err error)
	Push(p av.Packet) error
	IsClosed() bool
	IsFresh() bool
	ShutDown()
}
