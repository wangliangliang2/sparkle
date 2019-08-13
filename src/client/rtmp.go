package client

import (
	"av"
	"buff"
	"chunkstream"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"log"
	"net"
	"protocol/flv/amf0"
	"time"
)

type Rtmp struct {
	conn net.Conn
	*buff.ReadWriter
	chunks         map[uint32]*chunkstream.ChunkStream
	writeChunkSize uint32
	readChunkSize  uint32
	uri            string
	token          string
	tunnel         chan *av.Packet
	isPub          bool
	isOld          bool
	Ready          chan struct{}
}

func NewRtmpClient(conn net.Conn) *Rtmp {
	return &Rtmp{
		conn:           conn,
		ReadWriter:     buff.New(conn),
		chunks:         make(map[uint32]*chunkstream.ChunkStream),
		writeChunkSize: chunkstream.DefaultChunkSize,
		readChunkSize:  chunkstream.DefaultChunkSize,
		tunnel:         make(chan *av.Packet),
		Ready:          make(chan struct{}),
	}
}

func (R Rtmp) IsHls() (ret bool) {
	ret = false
	return
}

func (R *Rtmp) IsOld() (ret bool) {
	ret = R.isOld
	R.isOld = true
	return
}

func (R *Rtmp) IsPub() (ret bool) {
	ret = R.isPub
	return
}

func (R *Rtmp) Uri() (ret string) {
	ret = R.uri
	return
}

func (R *Rtmp) Token() (ret string) {
	ret = R.token
	return
}

func (R *Rtmp) GetPacket() (packet *av.Packet, ok bool) {
	packet, ok = <-R.tunnel
	return
}

func (R *Rtmp) ReceivePacket(packet *av.Packet) {
	chunk := packet.ToChunkStream()
	chunk.Write(R.ReadWriter, R.writeChunkSize)
}

func (R *Rtmp) Close() {
	if err := R.CloseConn(); err != nil {
		return
	}
	R.CloseChannel()
}

func (R *Rtmp) CloseConn() (err error) {
	err = R.conn.Close()
	return
}

func (R *Rtmp) CloseChannel() {
	close(R.tunnel)
	close(R.Ready)
}

func (R *Rtmp) Work() {
	R.handshake(prepareHandshakeData())
	R.readStream()
}

func (R *Rtmp) handshake(C1, C2, C0C1, S1, S2, S0S1S2 []byte) {
	if _, err := R.Read(C0C1); err != nil {
		R.Close()
		return
	}

	if ok := simpleHandshake(C1, S1, S2); !ok {
		if _, ok := complexHandshake(C1, S1, S2); !ok {
			R.Close()
			return
		}
	}

	if _, err := R.WriteAtOnce(S0S1S2); err != nil {
		R.Close()
		return
	}

	if _, err := R.Read(C2); err != nil {
		R.Close()
		return
	}
}

func (R *Rtmp) readStream() (err error) {
	for {
		chunk := &chunkstream.ChunkStream{}
		if chunk, err = R.getChunk(); err != nil {
			log.Println(R.token, err)
			R.Close()
			return
		}
		R.handleChunk(chunk)
	}
	return
}

func (R *Rtmp) getChunk() (current *chunkstream.ChunkStream, err error) {
	current = &chunkstream.ChunkStream{}
	for {
		var csid uint32
		if csid, err = chunkstream.PeekCSID(R.ReadWriter); err != nil {
			return
		}
		if old, ok := R.chunks[csid]; ok {
			current = old
		}

		var complete bool
		if complete, err = current.Read(R.ReadWriter, R.readChunkSize); err != nil {
			return
		}

		R.chunks[csid] = current
		if complete {
			break
		}
	}
	return
}

func (R *Rtmp) handleChunk(chunk *chunkstream.ChunkStream) {

	switch chunk.MsgTypeID {
	case chunkstream.MsgTypeIDSetChunkSize:
		chunksize := binary.BigEndian.Uint32(chunk.Data)
		R.readChunkSize = chunksize
	case chunkstream.MsgtypeIDCommandMsgAMF0:
		R.handleChunkCommand(amf0.NewAmf0FromBytes(chunk.Data))
	case chunkstream.MsgtypeIDAudioMsg, chunkstream.MsgtypeIDVideoMsg, chunkstream.MsgtypeIDDataMsgAMF0:
		R.tunnel <- av.NewPacket(chunk)

	}
}

func (R *Rtmp) handleChunkCommand(cmd *amf0.Amf0Cmd) {
	R.responseToCmd(cmd)
	R.makeUserInfo(cmd)
}

func (R *Rtmp) responseToCmd(cmd *amf0.Amf0Cmd) {
	switch cmd.CmdName {
	case chunkstream.NetConnCmdConnect:
		// R.setWriteChunkSize(60000)
		R.responseToCmdConnect(cmd)
	case chunkstream.NetConnCmdCreateStream:
		R.responseToCmdCreateStream(cmd)
	case chunkstream.NetStreamCmdsPublish:
		R.responseToCmdsPublish(cmd)
	case chunkstream.NetStreamCmdsPlay:
		R.responseToCmdsPlay(cmd)
	}
}

func (R *Rtmp) makeUserInfo(cmd *amf0.Amf0Cmd) {
	R.makeToken(cmd)
	R.makeNotice(cmd)
}

func (R *Rtmp) makeToken(cmd *amf0.Amf0Cmd) {
	switch cmd.CmdName {
	case chunkstream.NetConnCmdConnect:
		R.uri = cmd.Obj["app"].(string) + "/"
	case chunkstream.NetStreamCmdsPlay, chunkstream.NetStreamCmdsPublish:
		R.uri += cmd.Opts[0].(string)
		text := R.uri + time.Now().Format("2006-01-02 15:04:05")
		hasher := md5.New()
		hasher.Write([]byte(text))
		R.token = hex.EncodeToString(hasher.Sum(nil))
	}
}

func (R *Rtmp) makeNotice(cmd *amf0.Amf0Cmd) {
	if cmd.CmdName == chunkstream.NetStreamCmdsPublish {
		R.isPub = true
	}
	switch cmd.CmdName {
	case chunkstream.NetStreamCmdsPlay, chunkstream.NetStreamCmdsPublish:
		R.Ready <- struct{}{}
	}
}

func (R *Rtmp) setWriteChunkSize(chunkSize uint32) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, chunkSize)
	chunkstream.New(chunkstream.ProtocolCtlMsgSetChunkSizeTypeID, 2, 0, data).Write(R.ReadWriter, R.writeChunkSize)
	R.writeChunkSize = chunkSize
}

func (R *Rtmp) responseToCmdConnect(cmd *amf0.Amf0Cmd) {
	R.WriteCmd(chunkstream.NetConnCmdResponseYES, chunkstream.CSIDServerAmf0Response, chunkstream.UserControlStreamID, cmd.TransID, amf0.Amf0ConnectOpt0, amf0.Amf0ConnectOpt1)
}

func (R *Rtmp) responseToCmdCreateStream(cmd *amf0.Amf0Cmd) {
	R.WriteCmd(chunkstream.NetConnCmdResponseYES, chunkstream.CSIDServerAmf0Response, chunkstream.UserControlStreamID, cmd.TransID, nil, chunkstream.PublishStreamID)
}

func (R *Rtmp) responseToCmdsPublish(cmd *amf0.Amf0Cmd) {
	R.WriteCmd(chunkstream.NetStreamCmdsResponse, chunkstream.CSIDServerAmf0Response, chunkstream.UserControlStreamID, cmd.TransID, nil, amf0.AMf0Publish)
}

func (R *Rtmp) responseToCmdsPlay(cmd *amf0.Amf0Cmd) {
	R.setStreamBegin()
	R.startPlay(cmd)
}

func (R *Rtmp) setStreamBegin() {
	data := make([]byte, 6)
	eventType := data[:2]
	binary.BigEndian.PutUint16(eventType, chunkstream.UserControlStreamBegin)
	eventData := data[2:]
	binary.BigEndian.PutUint32(eventData, chunkstream.PublishStreamID)
	chunk := chunkstream.New(chunkstream.MsgtypeIDUserControl, chunkstream.CSIDUserControl, chunkstream.UserControlStreamID, data)
	chunk.Write(R.ReadWriter, R.writeChunkSize)
}

func (R *Rtmp) startPlay(cmd *amf0.Amf0Cmd) {
	R.WriteCmd(chunkstream.NetStreamCmdsResponse, chunkstream.CSIDServerAmf0Cmd, chunkstream.PublishStreamID, cmd.TransID, nil, amf0.AMf0Play)
}

func (R *Rtmp) WriteCmd(cmdName string, csid, streamid uint32, transid float64, cmd amf0.Amf0Map, opt ...interface{}) {
	amf0Cmd := amf0.NewAmf0(cmdName, transid, cmd, opt...)
	chunk := chunkstream.New(chunkstream.MsgtypeIDCommandMsgAMF0, csid, streamid, amf0Cmd.Encode())
	chunk.Write(R.ReadWriter, R.writeChunkSize)
}
