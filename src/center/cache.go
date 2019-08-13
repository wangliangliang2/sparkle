package center

import (
	"av"
	"client"
)

type Cache struct {
	Metadata *av.Packet
	VideoSeq *av.Packet
	AudioSeq *av.Packet
	Gop      []*av.Packet
}

func NewCache() *Cache {
	return &Cache{
		Gop: make([]*av.Packet, 0, 1024),
	}
}

func (C *Cache) Save(p *av.Packet) {
	switch {
	case p.IsMetadata:
		C.Metadata = p
	case p.IsAudioSequence():
		C.AudioSeq = p
	case p.IsVideoSequence():
		C.VideoSeq = p
	case p.IsPureAudioData(), p.IsPureVideoData():
		if p.IsVideoKeyFrame() {
			C.Gop = make([]*av.Packet, 0, 1024)
		}
		C.Gop = append(C.Gop, p)
	}
}

func (C *Cache) Send(cli client.Client) {
	if C.Metadata != nil {
		cli.ReceivePacket(C.Metadata)
	}
	if C.VideoSeq != nil {
		cli.ReceivePacket(C.VideoSeq)
	}
	if C.AudioSeq != nil {
		cli.ReceivePacket(C.AudioSeq)
	}
	for _, cache := range C.Gop {
		cli.ReceivePacket(cache)
	}
}
