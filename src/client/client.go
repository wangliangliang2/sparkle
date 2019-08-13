package client

import "av"

type Client interface {
	Uri() string
	Token() string
	GetPacket() (packet *av.Packet, ok bool)
	ReceivePacket(packet *av.Packet)
	Close()
	IsOld() bool
	IsHls() bool
}
