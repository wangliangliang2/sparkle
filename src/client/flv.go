package client

import (
	"av"
	"bits"
	"net/http"
	"protocol/flv"
)

type FlvClient struct {
	Data              chan av.Packet
	WebHandle         http.ResponseWriter
	isClosed, isFresh bool
}

func NewFlvClient(w http.ResponseWriter) *FlvClient {
	return &FlvClient{
		Data:      make(chan av.Packet, 1024),
		isFresh:   true,
		WebHandle: w,
	}
}

func (F *FlvClient) Pull() (p av.Packet, err error) {
	return
}

func (F *FlvClient) Push(p av.Packet) error {
	F.Data <- p
	return nil
}
func (F *FlvClient) IsClosed() bool {
	return F.isClosed
}

func (F *FlvClient) IsFresh() bool {
	defer func() { F.isFresh = false }()
	return F.isFresh
}

func (F *FlvClient) ShutDown() {
	close(F.Data)
}

func (F *FlvClient) Play() {
	F.Write([]byte{0x46, 0x4C, 0x56, 0x01, 0x05, 0x00, 0x00, 0x00, 0x09, 0x00, 0x00, 0x00, 0x00})
	for {
		p, ok := <-F.Data
		if !ok {
			break
		}
		switch {
		case p.IsMetadata:
			F.Write([]byte{flv.TAG_SCRIPTDATAAMF0})
		case p.IsAudio:
			F.Write([]byte{flv.TAG_AUDIO})
		case p.IsVideo:
			F.Write([]byte{flv.TAG_VIDEO})
		}
		F.Write(bits.PutU24BE(uint32(len(p.Data))))
		if p.TimeStamp >= 0xFFFFFF {
			F.Write(bits.PutU32BE(p.TimeStamp))
		} else {
			F.Write(bits.PutU24BE(p.TimeStamp))
			F.Write([]byte{0x00})
		}
		F.Write(bits.PutU24BE(p.StreamID))
		F.Write(p.Data)
		// tag type | tag data size | timestamp | extendTimestamp | streamid |tag data
		//   1byte      3bytes         3bytes       1byte           3bytes      nbytes
		F.Write(bits.PutU32BE(uint32(len(p.Data) + 11)))

	}
}

func (F *FlvClient) Write(data []byte) {
	if F.isClosed {
		return
	}
	if _, err := F.WebHandle.Write(data); err != nil {
		// if user close then tell program remove.
		F.isClosed = true
	}
}
