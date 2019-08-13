package tag

import (
	"bytes"
)

const (
	Audio          = 8
	Video          = 9
	ScriptDataAmf0 = 0x12
	ScriptDataAmf3 = 0xf
)

const (
	SoundMp3                 = 2
	SoundNellyMoser16KhzMono = 5
	SoundNellyMoser8KhzMono  = 5
	SoundNellyMoser          = 6
	SoundALaw                = 7
	SoundMuLaw               = 8
	SoundAac                 = 10
	SoundSpeex               = 11

	Sound5_5Khz = 0
	Sound11Khz  = 1
	Sound22Khz  = 2
	Sound44Khz  = 3

	Sound8Bit  = 0
	Sound16Bit = 1

	SoundMono   = 0
	SoundStereo = 1

	AacSeqHeader = 0
	AacRaw       = 1

	AvcSeqHeader = 0
	AvcNalu      = 1
	AvcEos       = 2
	FrameKey     = 1
	FrameInter   = 2
	Video_H264   = 7
)

type Tag struct {
	TagHeader
	TagBody
	MediaInfo
}

type TagHeader struct {
	TypeID    uint8
	DataSize  uint32
	Timestamp uint32
	// streamid  uint32
}

type TagBody struct {
	Data []byte
}

type MediaInfo struct {
	FrameType       uint8
	CodecID         uint8
	AvcPacketType   uint8
	CompositionTime uint32
	SoundFormat     uint8
	SoundRate       uint8
	SoundSize       uint8
	SoundType       uint8
	AacPacketType   uint8
}

var FlvFileHeader = []byte{0x46, 0x4C, 0x56, 0x01, 0x05, 0x00, 0x00, 0x00, 0x09, 0x00, 0x00, 0x00, 0x00}

var streamid = []byte{0x00, 0x00, 0x00}

func New(typeid uint8, timestamp uint32, data []byte) (t Tag) {
	t = Tag{}
	t.TypeID = typeid
	t.Timestamp = timestamp
	t.Data = data
	switch t.TypeID {
	case Audio:
		t.ParseAudioInfo(data)
	case Video:
		t.ParseVideoInfo(data)
	}
	return
}

func (t *Tag) ParseAudioInfo(b []byte) (err error) {
	flags := b[0]
	t.SoundFormat = flags >> 4
	t.SoundRate = (flags >> 2) & 0x3
	t.SoundSize = (flags >> 1) & 0x1
	t.SoundType = flags & 0x1
	if t.SoundFormat == SoundAac {
		t.AacPacketType = b[1]
	}
	return
}

func (t *Tag) ParseVideoInfo(b []byte) (err error) {
	flags := b[0]
	t.FrameType = flags >> 4
	t.CodecID = flags & 0xf
	if t.CodecID == Video_H264 {
		t.AvcPacketType = b[1]
		t.CompositionTime = uint32(b[2])<<16 + uint32(b[3])<<8 + uint32(b[4])
	}
	return
}

func (t Tag) IsAudioSequence() bool {
	return t.SoundFormat == SoundAac && t.AacPacketType == AvcSeqHeader
}

func (t Tag) IsVideoSequence() bool {
	return t.FrameType == FrameKey && t.AvcPacketType == AvcSeqHeader
}

func (t Tag) IsPureAudioData() bool {
	return t.AacPacketType != AacSeqHeader
}

func (t Tag) IsPureVideoData() bool {
	return t.AvcPacketType != AvcSeqHeader
}

func (t Tag) IsVideoKeyFrame() bool {
	return t.FrameType == FrameKey
}

func (t Tag) ToFlvTag() (ret []byte) {
	var tmp bytes.Buffer
	tmp.WriteByte(t.getTagtypeId())
	tmp.Write(t.getTagDataLength())
	tmp.Write(t.getTagTimestamp())
	tmp.Write(t.getTagStreamId())
	tmp.Write(t.getTagData())
	tmp.Write(t.getTagLength())
	ret = tmp.Bytes()
	return
}

func (t Tag) getTagtypeId() (ret byte) {
	ret = t.TypeID
	return
}

func (t Tag) getTagDataLength() (ret []byte) {
	ret = t.WriteBE(len(t.Data), 3)
	return
}

func (t Tag) getTagTimestamp() (ret []byte) {
	timestamp := make([]byte, 4)
	timestamp[0] = byte(t.Timestamp >> 16)
	timestamp[1] = byte(t.Timestamp >> 8)
	timestamp[2] = byte(t.Timestamp)
	timestamp[3] = byte(t.Timestamp >> 24)
	ret = timestamp
	return
}

func (t Tag) getTagStreamId() (ret []byte) {
	ret = streamid
	return
}

func (t Tag) getTagData() (ret []byte) {
	ret = t.Data
	return
}

func (t Tag) getTagLength() (ret []byte) {
	ret = t.WriteBE(len(t.Data)+11, 4)
	return
}

func (t Tag) WriteBE(val int, n int) (ret []byte) {
	ret = make([]byte, n)
	for index := 0; index < n; index++ {
		ret[index] = byte(val >> uint((n-index-1)*8))
	}
	return
}
