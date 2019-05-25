package chunk

import (
	"bits"
	"buff"
	"log"
	"protocol/flv"
)

const (
	_ = iota
	SetChunkSizeTypeID
	AbortMessageTypeID
	AckTypeID
	UserControlMessagesTypeID
	WindowAckSizeTypeID
	SetPeerBandwidthTypeID
)

type ChunkStream struct {
	Format    uint8
	TmpFormat uint8
	CSID      uint32

	Timestamp uint32
	Length    uint32
	TypeID    uint8
	StreamID  uint32

	extend    bool
	Complete  bool
	remain    uint32
	TimeDelta uint32
	index     uint32
	Data      []byte
}

func (C *ChunkStream) ReadBasicHeader(rw *buff.ReadWriter) {
	header, _ := rw.ReadUint32LE(1)
	C.TmpFormat = uint8(header >> 6)
	cstype := uint8(header) & 0x3f
	switch cstype {
	default: // Basic Header take 1 byte
		C.CSID = uint32(cstype)
	case 0x00: // Basic Header take 2 bytes
		csid, _ := rw.ReadUint32LE(1)
		C.CSID = csid + 64
	case 0x3f: // Basic Header take 3 bytes
		csid, _ := rw.ReadUint32LE(2)
		C.CSID = csid + 64
	}
}

func (C *ChunkStream) ReadChunk(rw *buff.ReadWriter, chunkSize uint32) (err error) {
	switch C.TmpFormat {
	case 0:
		C.Format = C.TmpFormat
		C.Timestamp, _ = rw.ReadUint32BE(3)
		C.Length, _ = rw.ReadUint32BE(3)
		typeid, _ := rw.ReadUint32BE(1)
		C.TypeID = uint8(typeid)
		C.StreamID, _ = rw.ReadUint32LE(4)
		if C.Timestamp == 0xffffff {
			C.Timestamp, _ = rw.ReadUint32BE(4)
			C.extend = true
		} else {
			C.extend = false
		}
		C.reset()
	case 1:
		C.Format = C.TmpFormat
		C.TimeDelta, _ = rw.ReadUint32BE(3)
		C.Length, _ = rw.ReadUint32BE(3)
		typeid, _ := rw.ReadUint32BE(1)
		C.TypeID = uint8(typeid)
		if C.TimeDelta == 0xffffff {
			C.TimeDelta, _ = rw.ReadUint32BE(4)
			C.extend = true
		} else {
			C.extend = false
		}
		C.Timestamp += C.TimeDelta
		C.reset()
	case 2:
		C.Format = C.TmpFormat
		C.TimeDelta, _ = rw.ReadUint32BE(3)
		if C.TimeDelta == 0xffffff {
			C.TimeDelta, _ = rw.ReadUint32BE(4)
			C.extend = true
		} else {
			C.extend = false
		}
		C.Timestamp += C.TimeDelta
		C.reset()
	case 3:
		switch C.Format {
		case 0:
			if C.extend {
				rw.Discard(4)
			}
		case 1, 2:
			if C.extend {
				C.TimeDelta, _ = rw.ReadUint32BE(4)
			}
			C.Timestamp += C.TimeDelta
		}
		if C.remain == 0 {
			C.reset()
		}
	}

	size := C.remain
	if size > chunkSize {
		size = chunkSize
	}

	buf := C.Data[C.index : C.index+size]
	if _, err = rw.Read(buf); err != nil {
		return
	}

	C.index += size
	C.remain -= size
	if C.remain == 0 {
		C.Complete = true
	}
	return
}

func (C *ChunkStream) reset() {
	C.Data = make([]byte, C.Length)
	C.Complete = false
	C.remain = C.Length
	C.index = 0
}

func (C *ChunkStream) WriteChunk(rw *buff.ReadWriter, chunkSize uint32) (err error) {
	switch C.TypeID {
	case flv.TAG_AUDIO:
		C.CSID = 4
	case flv.TAG_VIDEO, flv.TAG_SCRIPTDATAAMF0, flv.TAG_SCRIPTDATAAMF3:
		C.CSID = 6
	}
	totalLength := uint32(0)
	numChunks := C.Length / chunkSize
	for i := uint32(0); i <= numChunks; i++ {
		if totalLength == C.Length {
			break
		}
		if i == 0 {
			C.Format = 0
		} else {
			C.Format = 3
		}
		C.writeHeader(rw)
		start := i * chunkSize
		inc := chunkSize
		if C.Length-start <= chunkSize {
			inc = C.Length - start
		}
		totalLength += inc
		end := start + inc
		buf := C.Data[start:end]
		if _, err = rw.Write(buf); err != nil {
			log.Println(err)
			return
		}
		rw.Flush()
	}
	return
}

func (C *ChunkStream) writeHeader(rw *buff.ReadWriter) {
	format := uint32(C.Format) << 6
	switch {
	case C.CSID < 64:
		format |= C.CSID
		rw.WriteUint32LE(format, 1)
	case C.CSID-64 < 256:
		format |= 0x00
		rw.WriteUint32LE(format, 1)
		rw.WriteUint32LE(C.CSID-64, 1)

	case C.CSID-64 < 65536:
		format |= 0x3f
		rw.WriteUint32LE(format, 1)
		rw.WriteUint32LE(C.CSID-64, 2)
	}

	if C.Format == 3 {
		goto END
	}
	if C.Timestamp >= 0xffffff {
		rw.WriteUint32BE(0xffffff, 3)
	} else {
		rw.WriteUint32BE(C.Timestamp, 3)
	}
	if C.Format == 2 {
		goto END
	}
	rw.WriteUint32BE(C.Length, 3)
	rw.WriteUint32BE(uint32(C.TypeID), 1)
	if C.Format == 1 {
		goto END
	}
	rw.WriteUint32LE(C.StreamID, 4)
END:
	if C.Timestamp >= 0xffffff {
		rw.WriteUint32BE(C.Timestamp, 4)
	}
	return
}

func NewAck(size uint32) ChunkStream {
	return controlChunk(AckTypeID, 4, size)
}

func NewSetChunkSize(size uint32) ChunkStream {
	return controlChunk(SetChunkSizeTypeID, 4, size)
}

func NewWindowAckSize(size uint32) ChunkStream {
	return controlChunk(WindowAckSizeTypeID, 4, size)
}

func NewSetPeerBandwidth(size uint32) ChunkStream {
	chunk := controlChunk(SetPeerBandwidthTypeID, 5, size)
	chunk.Data = append(chunk.Data, 2)
	return chunk
}

func controlChunk(id uint8, size, value uint32) ChunkStream {
	return ChunkStream{
		Format:    0,
		CSID:      2,
		Timestamp: 0,
		Length:    size,
		TypeID:    id,
		StreamID:  0,
		Data:      bits.PutU32BE(value),
	}
}
