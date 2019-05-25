package flv

import "bits"

const (
	TAG_AUDIO          = 8
	TAG_VIDEO          = 9
	TAG_SCRIPTDATAAMF0 = 0x12
	TAG_SCRIPTDATAAMF3 = 0xf
)

const (
	MetadatAMF0  = 0x12
	MetadataAMF3 = 0xf
)

const (
	SOUND_MP3                   = 2
	SOUND_NELLYMOSER_16KHZ_MONO = 4
	SOUND_NELLYMOSER_8KHZ_MONO  = 5
	SOUND_NELLYMOSER            = 6
	SOUND_ALAW                  = 7
	SOUND_MULAW                 = 8
	SOUND_AAC                   = 10
	SOUND_SPEEX                 = 11

	SOUND_5_5Khz = 0
	SOUND_11Khz  = 1
	SOUND_22Khz  = 2
	SOUND_44Khz  = 3

	SOUND_8BIT  = 0
	SOUND_16BIT = 1

	SOUND_MONO   = 0
	SOUND_STEREO = 1

	AAC_SEQHDR = 0
	AAC_RAW    = 1
)

const (
	AVC_SEQHDR = 0
	AVC_NALU   = 1
	AVC_EOS    = 2

	FRAME_KEY   = 1
	FRAME_INTER = 2

	VIDEO_H264 = 7
)

type Tag struct {
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

func (tag *Tag) ParseAudioHeader(b []byte) (err error) {
	flags := b[0]
	tag.SoundFormat = flags >> 4
	tag.SoundRate = (flags >> 2) & 0x3
	tag.SoundSize = (flags >> 1) & 0x1
	tag.SoundType = flags & 0x1
	if tag.SoundFormat == SOUND_AAC {
		tag.AacPacketType = b[1]
	}
	return
}

func (tag *Tag) ParseVideoHeader(b []byte) (err error) {

	flags := b[0]
	tag.FrameType = flags >> 4
	tag.CodecID = flags & 0xf
	if tag.CodecID == VIDEO_H264 {
		tag.AvcPacketType = b[1]
		tag.CompositionTime = bits.U24BE(b[2:5])
	}
	return
}
