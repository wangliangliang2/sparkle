package av

import (
	"chunkstream"
	"protocol/flv/tag"
)

type Packet struct {
	IsAudio    bool
	IsVideo    bool
	IsMetadata bool
	TimeDelta  uint32
	StreamID   uint32
	CSID       uint32
	tag.Tag
}

func NewPacket(chunk *chunkstream.ChunkStream) (p *Packet) {
	p = &Packet{
		Tag:        tag.New(chunk.MsgTypeID, chunk.Timestamp, chunk.Data),
		CSID:       chunk.CSID,
		StreamID:   chunk.MsgStreamID,
		TimeDelta:  chunk.Timedelta,
		IsAudio:    chunk.MsgTypeID == chunkstream.MsgtypeIDAudioMsg,
		IsVideo:    chunk.MsgTypeID == chunkstream.MsgtypeIDVideoMsg,
		IsMetadata: chunk.MsgTypeID == chunkstream.MsgtypeIDDataMsgAMF0,
	}
	return
}

func (P *Packet) ToChunkStream() (chunk *chunkstream.ChunkStream) {
	chunk = &chunkstream.ChunkStream{}
	chunk.Data = P.Data
	chunk.MsgLen = uint32(len(P.Data))
	chunk.MsgStreamID = P.StreamID
	chunk.Timestamp = P.Timestamp
	chunk.CSID = P.CSID
	chunk.MsgTypeID = P.TypeID
	return
}
