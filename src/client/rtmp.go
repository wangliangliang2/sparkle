package client

import (
	"av"
	"bits"
	"buff"
	"bytes"
	"chunk"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"net"
	"protocol/flv"
)

var (
	FMSKey = []byte{
		0x47, 0x65, 0x6e, 0x75, 0x69, 0x6e, 0x65, 0x20,
		0x41, 0x64, 0x6f, 0x62, 0x65, 0x20, 0x46, 0x6c,
		0x61, 0x73, 0x68, 0x20, 0x4d, 0x65, 0x64, 0x69,
		0x61, 0x20, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72,
		0x20, 0x30, 0x30, 0x31, // Genuine Adobe Flash Media Server 001
		0xf0, 0xee, 0xc2, 0x4a, 0x80, 0x68, 0xbe, 0xe8,
		0x2e, 0x00, 0xd0, 0xd1, 0x02, 0x9e, 0x7e, 0x57,
		0x6e, 0xec, 0x5d, 0x2d, 0x29, 0x80, 0x6f, 0xab,
		0x93, 0xb8, 0xe6, 0x36, 0xcf, 0xeb, 0x31, 0xae,
	}

	FPKey = []byte{
		0x47, 0x65, 0x6E, 0x75, 0x69, 0x6E, 0x65, 0x20,
		0x41, 0x64, 0x6F, 0x62, 0x65, 0x20, 0x46, 0x6C,
		0x61, 0x73, 0x68, 0x20, 0x50, 0x6C, 0x61, 0x79,
		0x65, 0x72, 0x20, 0x30, 0x30, 0x31, // Genuine Adobe Flash Player 001
		0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8,
		0x2E, 0x00, 0xD0, 0xD1, 0x02, 0x9E, 0x7E, 0x57,
		0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
		0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	}
)

const (
	MsgTypeIDSetChunkSize              = 1
	MsgTypeIDAbortMessage              = 2
	MsgTypeIDACK                       = 3
	MsgTypeIDWindowAcknowledgementSize = 5
	MsgTypeIDSetPeerBandwidth          = 6
	MsgtypeIDUserControl               = 4
	MsgtypeIDCommandMsgAMF0            = 20
	MsgtypeIDCommandMsgAMF3            = 17
	MsgtypeIDDataMsgAMF0               = 18
	MsgtypeIDDataMsgAMF3               = 15
	MsgtypeIDVideoMsg                  = 9
	MsgtypeIDAudioMsg                  = 8
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

type StreamInfo struct {
	App            string
	Name           string
	IsPub          bool
	csid, streamID uint32
}

type RtmpClient struct {
	isClosed, isFresh                  bool
	Conn                               net.Conn
	io                                 *buff.ReadWriter
	chunks                             map[uint32]*chunk.ChunkStream
	chunkSize, remoteChunkSize         uint32
	windowAckSize, remoteWindowAckSize uint32
	ackReceived                        uint32
	limitType                          uint32
	StreamInfo                         StreamInfo
}

func NewRtmpClient(conn net.Conn) *RtmpClient {
	return &RtmpClient{
		chunks:              make(map[uint32]*chunk.ChunkStream),
		isFresh:             true,
		Conn:                conn,
		io:                  buff.New(conn),
		chunkSize:           128,
		remoteChunkSize:     128,
		windowAckSize:       2500000,
		remoteWindowAckSize: 2500000,
		limitType:           2,
	}
}

func (R *RtmpClient) Pull() (p av.Packet, err error) {
	var cs chunk.ChunkStream
	err = R.Read(&cs)
	if err != nil {
		return av.Packet{}, err
	}
	_, err = R.handleMsg(&cs)
	if err != nil {
		return av.Packet{}, err
	}
	p = av.NewPacket(&cs)
	return p, nil
}

func (R *RtmpClient) Push(p av.Packet) error {
	cs := p.TransferToChunkStream()
	cs.WriteChunk(R.io, R.chunkSize)
	R.io.Flush()
	return nil
}

func (R *RtmpClient) IsClosed() bool {
	return R.isClosed
}

func (R *RtmpClient) IsFresh() bool {
	defer func() { R.isFresh = false }()
	return R.isFresh
}

func (R *RtmpClient) ShutDown() {
	R.isClosed = true
	R.Conn.Close()
	R.chunks = nil
}

func (R *RtmpClient) ReceiveData() {

	if R.StreamInfo.IsPub {

		R.sendCmd(R.StreamInfo.csid, R.StreamInfo.streamID, "onStatus", 0, nil, flv.AMFMap{
			"level":       "status",
			"code":        "NetStream.Publish.Start",
			"description": "Start publishing",
		})
	} else {
		R.responsePlay(R.StreamInfo.csid, R.StreamInfo.streamID)
		R.Start()
	}
}

func (R *RtmpClient) Handshake() (err error) {
	var random [(1 + 1536*2) * 2]byte

	C0C1C2 := random[:1536*2+1]
	C1 := C0C1C2[1 : 1536+1]
	C0C1 := C0C1C2[:1536+1]
	C2 := C0C1C2[1536+1:]

	S0S1S2 := random[1536*2+1:]
	S0 := S0S1S2[:1]
	S1 := S0S1S2[1 : 1536+1]
	S2 := S0S1S2[1536+1:]

	S0[0] = 3 // server rtmp version

	if _, err = R.io.Read(C0C1); err != nil {
		return
	}

	clientVersion := bits.U32BE(C1[4:8])
	if clientVersion == 0 {
		// simple handshake
		copy(S1, C1)
		copy(S2, C1)
	} else {
		// complex handshake
		if digestData, ok := getDigestData(C1); !ok {
			return
		} else {
			clientTime := C1[0:4]
			makeS1(S1, clientTime)
			makeS2(S2, digestData)
		}
	}
	if _, err = R.io.Write(S0S1S2); err != nil {
		return
	}
	if err = R.io.Flush(); err != nil {
		return
	}

	if _, err = R.io.Read(C2); err != nil {
		return
	}

	return
}

func getDigestData(C1 []byte) (digestData []byte, ok bool) {

	if digestData, ok = testDigest(C1, 772); !ok {
		digestData, ok = testDigest(C1, 8)
	}
	return
}

func testDigest(data []byte, base int) (digestData []byte, ok bool) {
	digestData, c1Part1, c1Part2 := findDigest(data, base)
	digest := calcDigest(c1Part1, c1Part2, FPKey[:30])
	ok = bytes.Compare(digestData, digest) == 0
	return
}

func findDigest(data []byte, base int) (digestData, part1, part2 []byte) {
	offsetBytes := data[base : base+4]
	var offset int

	for _, val := range offsetBytes {
		offset += int(val)
	}
	totalOffset := base + 4 + (offset % 728)

	digestData = data[totalOffset : totalOffset+32]
	part1 = data[:totalOffset]
	part2 = data[totalOffset+32:]
	return
}

func calcDigest(part1, part2, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(part1)
	h.Write(part2)
	return h.Sum(nil)
}

func makeS1(S1, timestamp []byte) {
	rand.Read(S1)
	copy(S1[:4], timestamp)
	copy(S1[4:8], []byte{0x04, 0x05, 0x00, 0x01}) //server version, not why just do it
	_, part1, part2 := findDigest(S1, 8)
	digestData := calcDigest(part1, part2, FMSKey[:36])
	var offset uint32
	for _, val := range S1[8 : 8+4] {
		offset += uint32(val)
	}
	totalOffset := 8 + 4 + (offset % 728)
	copy(S1[totalOffset:totalOffset+32], digestData)
}

func makeS2(S2, digestData []byte) {
	tempkey := calcDigest(digestData, nil, FMSKey)
	rand.Read(S2)
	length := len(S2) - 32
	digest := calcDigest(S2[:length], nil, tempkey)
	copy(S2[length:], digest)
}

func (R *RtmpClient) Start() (programName string, isPub bool) {
	var cs chunk.ChunkStream
	for {
		err := R.Read(&cs)
		if err != nil {
			R.isClosed = true
			break
		}
		stop, err := R.handleMsg(&cs)
		if err != nil || stop {
			break
		}
	}
	return R.StreamInfo.App + "/" + R.StreamInfo.Name, R.StreamInfo.IsPub
}

func (R *RtmpClient) Read(c *chunk.ChunkStream) (err error) {
	var cs chunk.ChunkStream
	for {
		cs.ReadBasicHeader(R.io)
		if tmp, ok := R.chunks[cs.CSID]; ok {
			tmp.TmpFormat = cs.TmpFormat
			cs = *tmp
		}
		if err = cs.ReadChunk(R.io, R.remoteChunkSize); err != nil {
			return
		}

		R.chunks[cs.CSID] = &cs

		if cs.Complete {
			*c = cs
			break
		}
	}
	R.ack(c.Length)
	return
}

func (R *RtmpClient) ack(size uint32) {
	R.ackReceived += size
	if R.ackReceived >= R.remoteWindowAckSize {
		cs := chunk.NewAck(R.ackReceived)
		cs.WriteChunk(R.io, R.chunkSize)
		R.io.Flush()
		R.ackReceived = 0
	}
}

func (R *RtmpClient) handleMsg(cs *chunk.ChunkStream) (stop bool, err error) {
	switch cs.TypeID {
	case MsgTypeIDSetChunkSize:
		R.remoteChunkSize = bits.U32BE(cs.Data)
	case MsgTypeIDAbortMessage, MsgTypeIDACK, MsgTypeIDWindowAcknowledgementSize, MsgtypeIDUserControl:
		R.remoteWindowAckSize = bits.U32BE(cs.Data)
	case MsgTypeIDSetPeerBandwidth:
		windowsize := bits.U32BE(cs.Data[:4])
		limitType := uint32(cs.Data[4])
		resend := R.windowAckSize == windowsize
		switch limitType {
		case 0:
			R.windowAckSize = windowsize
		case 1:
			if R.windowAckSize > windowsize {
				R.windowAckSize = windowsize
			}
		case 2:
			if R.limitType == 0 {
				R.windowAckSize = windowsize
			}
		}
		R.limitType = limitType
		if resend {
			cs := chunk.NewWindowAckSize(R.windowAckSize)
			cs.WriteChunk(R.io, R.chunkSize)
			R.io.Flush()
		}
	case MsgtypeIDCommandMsgAMF3, MsgtypeIDCommandMsgAMF0:
		if cs.TypeID == MsgtypeIDCommandMsgAMF3 {
			cs.Data = cs.Data[1:]
		}
		val := flv.ParserCommandMsgAMF0(cs.Data)
		stop, err = R.handleCommandMsg(cs.CSID, cs.StreamID, val)
	}
	return
}

func (R *RtmpClient) handleCommandMsg(csid, streamID uint32, arg []interface{}) (stop bool, err error) {

	commandname := arg[0].(string)
	switch commandname {
	case "connect":
		info := arg[2].(flv.AMFMap)
		R.StreamInfo = StreamInfo{
			App: info["app"].(string),
		}
		R.responseConnect(csid, streamID)
	case "createStream":
		publishStreamID := 1
		transactionID := arg[1].(float64)
		R.sendCmd(csid, streamID, "_result", transactionID, nil, publishStreamID)
	case "publish":
		R.StreamInfo.IsPub = true
	case "play":
		R.StreamInfo.IsPub = false
	case "releaseStream", "FCPublish", "getStreamLength", "deleteStream", "FCUnpublish":
	case "seek":
		R.sendCmd(csid, streamID, "onStatus", 0, nil, flv.AMFMap{
			"level":       "status",
			"code":        "NetStream.Play.Start",
			"description": "Start playing stream.",
		})
	}
	if commandname == "publish" || commandname == "play" {
		R.StreamInfo.Name = arg[3].(string)
		R.StreamInfo.csid = csid
		R.StreamInfo.streamID = streamID
		stop = true
	}
	return
}

func (R *RtmpClient) responseConnect(csid, streamID uint32) {
	var cs chunk.ChunkStream
	cs = chunk.NewWindowAckSize(R.windowAckSize)
	cs.WriteChunk(R.io, R.chunkSize)

	cs = chunk.NewSetPeerBandwidth(R.remoteWindowAckSize)
	cs.WriteChunk(R.io, R.chunkSize)
	setchunkSize := uint32(60000)
	cs = chunk.NewSetChunkSize(setchunkSize)
	cs.WriteChunk(R.io, R.chunkSize)
	R.chunkSize = setchunkSize
	transactionID := 1
	R.sendCmd(csid, streamID, "_result", transactionID, flv.AMFMap{
		"fmtVer":       "FMS/3,0,1,123",
		"capabilities": 31,
	}, flv.AMFMap{
		"level":          "status",
		"code":           "NetConnection.Connect.Success",
		"description":    "Connection succeeded.",
		"objectEncoding": 0,
	})

}

func (R *RtmpClient) responsePlay(csid, streamID uint32) {
	R.SetStreamBegin()
	R.sendCmd(csid, streamID, "onStatus", 0, nil, flv.AMFMap{
		"level":       "status",
		"code":        "NetStream.Play.Start",
		"description": "Start playing stream.",
	})

}

func (R *RtmpClient) SetStreamBegin() {
	cs := chunk.ChunkStream{
		Format:   0,
		CSID:     2,
		TypeID:   4,
		StreamID: 0,
		Length:   6,
		Data:     make([]byte, 6),
	}
	copy(cs.Data[:2], bits.PutU16BE(UserControlStreamBegin))
	copy(cs.Data[2:], bits.PutU32BE(1))
	cs.WriteChunk(R.io, R.chunkSize)
}

func (R *RtmpClient) sendCmd(csid, msgSId uint32, args ...interface{}) (err error) {
	return R.AMF0Msg(csid, msgSId, MsgtypeIDCommandMsgAMF0, args...)
}

func (R *RtmpClient) AMF0Msg(csid, streamID uint32, msgTypeId uint8, args ...interface{}) (err error) {
	var size uint32
	body := make([]byte, 0)
	for _, arg := range args {
		size += uint32(flv.AMFValLen(arg))
		body = append(body, flv.AMFValFill(arg)...)
	}
	cs := chunk.ChunkStream{
		Format:    0,
		CSID:      csid,
		Timestamp: 0,
		TypeID:    msgTypeId,
		StreamID:  streamID,
		Length:    uint32(len(body)),
		Data:      body,
	}
	cs.WriteChunk(R.io, R.chunkSize)
	return R.io.Flush()
}
