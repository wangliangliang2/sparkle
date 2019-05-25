package av

import (
	"chunk"
	"protocol/flv"
)

type Packet struct {
	IsAudio    bool
	IsVideo    bool
	IsMetadata bool
	TimeStamp  uint32
	TimeDelta  uint32
	StreamID   uint32
	flv.Tag
	Data []byte
}

func NewPacket(cs *chunk.ChunkStream) Packet {
	tag := flv.Tag{}
	switch cs.TypeID {
	case flv.TAG_AUDIO:
		tag.ParseAudioHeader(cs.Data)
	case flv.TAG_VIDEO:
		tag.ParseVideoHeader(cs.Data)
	}
	p := Packet{
		IsAudio:    cs.TypeID == flv.TAG_AUDIO,
		IsVideo:    cs.TypeID == flv.TAG_VIDEO,
		IsMetadata: cs.TypeID == flv.TAG_SCRIPTDATAAMF0 || cs.TypeID == flv.TAG_SCRIPTDATAAMF3,
		StreamID:   cs.StreamID,
		Data:       cs.Data,
		TimeStamp:  cs.Timestamp,
		TimeDelta:  cs.TimeDelta,
		Tag:        tag,
	}

	return p
}

func (P *Packet) TransferToChunkStream() *chunk.ChunkStream {
	var cs chunk.ChunkStream
	cs.Data = P.Data
	cs.Length = uint32(len(P.Data))
	cs.StreamID = P.StreamID
	cs.Timestamp = P.TimeStamp

	switch {
	case P.IsVideo:
		cs.TypeID = flv.TAG_VIDEO
	case P.IsAudio:
		cs.TypeID = flv.TAG_AUDIO
	case P.IsMetadata:
		cs.TypeID = flv.TAG_SCRIPTDATAAMF0
	}
	return &cs
}

func (P *Packet) IsAudioSequence() bool {
	return P.SoundFormat == flv.SOUND_AAC && P.AacPacketType == flv.AAC_SEQHDR
}

func (P *Packet) IsVideoSequence() bool {
	return P.FrameType == flv.FRAME_KEY && P.AvcPacketType == flv.AVC_SEQHDR
}

func (P *Packet) IsPureAudioData() bool {
	return P.IsAudio && P.AacPacketType != flv.AAC_SEQHDR
}

func (P *Packet) IsPureVideoData() bool {
	return P.IsVideo && P.AvcPacketType != flv.AVC_SEQHDR
}

func (P *Packet) IsVideoKeyFrame() bool {
	return P.FrameType == flv.FRAME_KEY
}
