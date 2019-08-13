package chunkstream

import (
	"buff"
)

type BasicHeader struct {
	ChunkType    uint8
	TmpChunkType uint8
	CSID         uint32
}
type MessageHeader struct {
	Timestamp   uint32
	Timedelta   uint32
	MsgLen      uint32
	MsgTypeID   uint8
	MsgStreamID uint32
}
type ChunkStream struct {
	BasicHeader
	MessageHeader
	extend       bool
	complete     bool
	msgLenRemain uint32
	index        uint32
	Data         []byte
}

const (
	ChunkTypeZero = iota
	ChunkTypeOne
	ChunkTypeTwo
	ChunkTypeThree
)

const (
	_ = iota
	ProtocolCtlMsgSetChunkSizeTypeID
	ProtocolCtlMsgAbortMessageTypeID
	ProtocolCtlMsgAckTypeID
	ProtocolCtlMsgUserControlMessagesTypeID
	ProtocolCtlMsgWindowAckSizeTypeID
	ProtocolCtlMsgSetPeerBandwidthTypeID
)

const (
	UserControlStreamBegin = iota
	UserControlStreamEOF
	UserControlStreamDry
	UserControlSetBufferLength
	UserControlStreamIsRecorded
	UserControlPingRequest
	UserControlPingResponse
)

const (
	MsgTypeIDSetChunkSize              = 1
	MsgTypeIDAbortMessage              = 2
	MsgTypeIDACK                       = 3
	MsgtypeIDUserControl               = 4
	MsgTypeIDWindowAcknowledgementSize = 5
	MsgTypeIDSetPeerBandwidth          = 6
	MsgtypeIDAudioMsg                  = 8
	MsgtypeIDVideoMsg                  = 9
	MsgtypeIDDataMsgAMF3               = 15
	MsgtypeIDCommandMsgAMF3            = 17
	MsgtypeIDDataMsgAMF0               = 18
	MsgtypeIDCommandMsgAMF0            = 20
)
const (
	CSIDUserControl        = 2
	CSIDServerAmf0Cmd      = 5
	CSIDServerAmf0Response = 3
)

const (
	PublishStreamID     = 1
	UserControlStreamID = 0
)

const (
	NetConnCmdConnect      = "connect"
	NetConnCmdCreateStream = "createStream"
	NetConnCmdResponseYES  = "_result"
	NetConnCmdResponseNO   = "_error"
)

const (
	NetStreamCmdsPublish      = "publish"
	NetStreamCmdsPlay         = "play"
	NetStreamCmdsDeleteStream = "deleteStream"
	NetStreamCmdsResponse     = "onStatus"
)

const (
	MaxTimestamp     = 0xffffff
	DefaultChunkSize = 128
)

func New(msgTypeID uint8, csid, msgStreamID uint32, data []byte) (ret *ChunkStream) {
	ret = &ChunkStream{}
	ret.CSID = csid
	ret.MsgLen = uint32(len(data))
	ret.MsgTypeID = msgTypeID
	ret.MsgStreamID = msgStreamID
	ret.Data = data
	return
}

func PeekCSID(r *buff.ReadWriter) (csid uint32, err error) {
	var data []byte
	if data, err = r.Peeks(3); err != nil {
		return
	}
	cstype := uint8(data[0] & 0x3f)
	switch cstype {
	default: // Basic Header take 1 byte
		csid = uint32(cstype)
	case 0x00: // Basic Header take 2 bytes
		csid = uint32(data[1]) + 64
	case 0x3f: // Basic Header take 3 bytes
		csid = uint32(data[2])*256 + uint32(data[1]) + 64
	}
	return
}

func (C *ChunkStream) Read(r *buff.ReadWriter, chunkSize uint32) (complete bool, err error) {
	if err = C.ReadHeader(r); err != nil {
		return
	}
	C.clearCache()
	if complete, err = C.ReadBody(r, chunkSize); err != nil {
		return
	}
	return
}

func (C *ChunkStream) ReadHeader(r *buff.ReadWriter) (err error) {
	if err = C.readBasicHeader(r); err != nil {
		return
	}
	if err = C.readMessageHeader(r); err != nil {
		return
	}
	if err = C.readExtTimestamp(r); err != nil {
		return
	}
	return
}
func (C *ChunkStream) ReadBody(r *buff.ReadWriter, chunkSize uint32) (complete bool, err error) {
	if err = C.readData(r, chunkSize); err != nil {
		return
	}
	complete = C.complete
	return
}

func (C *ChunkStream) readBasicHeader(r *buff.ReadWriter) (err error) {
	var header byte
	if header, err = r.OneByte(); err != nil {
		return
	}
	C.ChunkType = uint8(header >> 6)
	cstype := uint8(header & 0x3f)
	switch cstype {
	default: // Basic Header take 1 byte
		C.CSID = uint32(cstype)
	case 0x00: // Basic Header take 2 bytes
		if C.CSID, err = r.ReadUint32LE(1); err != nil {
			return
		}
		C.CSID += 64
	case 0x3f: // Basic Header take 3 bytes
		if C.CSID, err = r.ReadUint32LE(2); err != nil {
			return
		}
		C.CSID += 64
	}
	return
}

func (C *ChunkStream) readMessageHeader(r *buff.ReadWriter) (err error) {

	if err = C.readTimestamp(r); err != nil {
		return
	}
	if err = C.readMsgLen(r); err != nil {
		return
	}
	if err = C.readMsgTypeID(r); err != nil {
		return
	}
	if err = C.readMsgStreamID(r); err != nil {
		return
	}
	return
}

func (C *ChunkStream) readTimestamp(r *buff.ReadWriter) (err error) {
	if C.ChunkType == ChunkTypeThree {
		return
	}
	var timestamp uint32
	if timestamp, err = r.ReadUint32BE(3); err != nil {
		return
	}
	if C.ChunkType == ChunkTypeZero {
		C.Timestamp = timestamp
	}
	if C.ChunkType == ChunkTypeOne || C.ChunkType == ChunkTypeTwo {
		C.Timedelta = timestamp
	}
	return
}

func (C *ChunkStream) readMsgLen(r *buff.ReadWriter) (err error) {
	if C.ChunkType == ChunkTypeZero || C.ChunkType == ChunkTypeOne {
		if C.MsgLen, err = r.ReadUint32BE(3); err != nil {
			return
		}
	}
	return
}

func (C *ChunkStream) readMsgTypeID(r *buff.ReadWriter) (err error) {
	if C.ChunkType == ChunkTypeZero || C.ChunkType == ChunkTypeOne {
		var msgid byte
		if msgid, err = r.OneByte(); err != nil {
			return
		}
		C.MsgTypeID = uint8(msgid)
	}
	return
}

func (C *ChunkStream) readMsgStreamID(r *buff.ReadWriter) (err error) {
	if C.ChunkType == ChunkTypeZero {
		if C.MsgStreamID, err = r.ReadUint32BE(4); err != nil {
			return
		}
	}
	return
}

func (C *ChunkStream) readExtTimestamp(r *buff.ReadWriter) (err error) {
	switch C.ChunkType {
	case ChunkTypeZero:
		if C.Timestamp == MaxTimestamp {
			if C.Timestamp, err = r.ReadUint32BE(4); err != nil {
				return
			}
			C.extend = true
		}
	case ChunkTypeOne, ChunkTypeTwo:
		if C.Timedelta == MaxTimestamp {
			if C.Timedelta, err = r.ReadUint32BE(4); err != nil {
				return
			}
			C.extend = true
		}
		C.Timestamp += C.Timedelta
	case ChunkTypeThree:
		switch C.TmpChunkType {
		case ChunkTypeZero, ChunkTypeThree:
			if C.extend {
				r.Discard(4)
			}
		case ChunkTypeOne, ChunkTypeTwo:
			if C.extend {
				if C.Timedelta, err = r.ReadUint32BE(4); err != nil {
					return
				}
			}
			C.Timestamp += C.Timedelta
		}
	}

	return
}

func (C *ChunkStream) clearCache() {
	if C.msgLenRemain == 0 {
		C.Data = make([]byte, C.MsgLen)
		C.index = 0
		C.msgLenRemain = C.MsgLen
		C.complete = false
	}

}

func (C *ChunkStream) readData(r *buff.ReadWriter, chunkSize uint32) (err error) {
	size := C.msgLenRemain
	if size > chunkSize {
		size = chunkSize
	}
	buf := C.Data[C.index : C.index+size]
	var n int
	if n, err = r.Read(buf); err != nil && n != int(size) {
		return
	}
	C.index += size
	C.msgLenRemain -= size
	if C.msgLenRemain == 0 {
		C.complete = true
	}
	C.TmpChunkType = C.ChunkType
	return
}

func (C *ChunkStream) Write(r *buff.ReadWriter, chunkSize uint32) {
	totalLength := uint32(0)
	numChunks := C.MsgLen / chunkSize
	for i := uint32(0); i <= numChunks; i++ {
		if totalLength == C.MsgLen {
			break
		}
		if i == 0 {
			C.ChunkType = 0
		} else {
			C.ChunkType = 3
		}
		C.WriteHeader(r)
		start := i * chunkSize
		inc := chunkSize
		if C.MsgLen-start <= chunkSize {
			inc = C.MsgLen - start
		}
		totalLength += inc
		end := start + inc
		data := C.Data[start:end]
		C.WriteBody(r, data)
	}
}

func (C *ChunkStream) WriteHeader(r *buff.ReadWriter) {
	C.writeBasicHeader(r)
	C.writeMsgHeader(r)
	C.writeExtTimestamp(r)
}

func (C *ChunkStream) writeBasicHeader(r *buff.ReadWriter) {
	chunkType := uint32(C.ChunkType) << 6
	switch {
	case C.CSID < 64:
		chunkType |= C.CSID
		r.WriteUint32LE(chunkType, 1)
	case C.CSID-64 < 256: //1<<8
		chunkType |= 0x00
		r.WriteUint32LE(chunkType, 1)
		r.WriteUint32LE(C.CSID-64, 1)
	case C.CSID-64 < 65536: //1<<16
		chunkType |= 0x3f
		r.WriteUint32LE(chunkType, 1)
		r.WriteUint32LE(C.CSID-64, 2)
	}
}

func (C *ChunkStream) writeMsgHeader(r *buff.ReadWriter) {
	C.writeTimestamp(r)
	C.writeMsgLen(r)
	C.writeMsgTypeID(r)
	C.writeMsgStreamID(r)
}

func (C *ChunkStream) writeTimestamp(r *buff.ReadWriter) (err error) {
	if C.ChunkType == ChunkTypeThree {
		return
	}
	if C.Timestamp >= MaxTimestamp {
		if err = r.WriteUint32BE(MaxTimestamp, 3); err != nil {
			return
		}
	} else {
		if err = r.WriteUint32BE(C.Timestamp, 3); err != nil {
			return
		}
	}
	return
}

func (C *ChunkStream) writeMsgLen(r *buff.ReadWriter) (err error) {
	if C.ChunkType == ChunkTypeZero || C.ChunkType == ChunkTypeOne {
		err = r.WriteUint32BE(uint32(len(C.Data)), 3)
	}
	return
}

func (C *ChunkStream) writeMsgTypeID(r *buff.ReadWriter) (err error) {
	if C.ChunkType == ChunkTypeZero || C.ChunkType == ChunkTypeOne {
		err = r.WriteByte(byte(C.MsgTypeID))
	}
	return
}

func (C *ChunkStream) writeMsgStreamID(r *buff.ReadWriter) (err error) {
	if C.ChunkType == ChunkTypeZero {
		err = r.WriteUint32LE(C.MsgStreamID, 4)
	}
	return
}

func (C *ChunkStream) writeExtTimestamp(r *buff.ReadWriter) (err error) {
	if C.Timestamp >= MaxTimestamp {
		err = r.WriteUint32BE(C.Timestamp, 4)
	}
	return
}

func (C *ChunkStream) WriteBody(r *buff.ReadWriter, data []byte) (n int, err error) {
	r.WriteAtOnce(data)
	return
}
