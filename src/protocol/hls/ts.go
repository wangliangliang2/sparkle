package hls

import (
	"av"
	"bytes"
	"encoding/binary"
)

const (
	TSPacketLen        = 188
	TSDefaultPacketLen = 184
	TSPCRPacketLen     = 176
	ADTSLen            = 7
	H264DefaultHZ      = 90
	VideoPID           = 0x100
	AudioPID           = 0x101
)

type MPEG4AudioCondig struct {
	SampleRateIndex byte
	ChannelIndex    byte
}

type TS struct {
	hasVideo, hasAudio bool
	AudioSeq, VideoSeq []byte
	MPEG4AudioCondig
}

func (T *TS) ReadSeq(p av.Packet) {
	switch {
	case p.IsAudioSequence():
		T.hasAudio = true
		T.AudioSeq = p.Data[2:] //move flv audio header
		T.getAudioSeqConfigInfo()
	case p.IsVideoSequence():
		T.hasVideo = true
		data := p.Data[5:] // move flv video header
		T.VideoSeq = convertVideoSeq(data)
	}
}

func (T *TS) ReadData(p av.Packet) []byte {
	var pesPacket []byte
	var pid, counter int
	ts := make([]byte, 0)
	data := make([]byte, len(p.Data))
	copy(data, p.Data)
	dts := int64(p.TimeStamp) * H264DefaultHZ
	switch {
	case p.IsPureAudioData():
		data := data[2:]
		audio := T.addADTS(data)
		pesPacket = addPES(audio, false, dts, 0)
		pid = AudioPID
		counter = AudioCount
	case p.IsPureVideoData():
		data := data[5:]
		video := convertVideo(data)
		if p.IsVideoKeyFrame() {
			video = T.addVideoSeq(video)
		}
		pts := dts + int64(p.CompositionTime)*H264DefaultHZ
		pesPacket = addPES(video, true, pts, dts)
		pid = VideoPID
		counter = VideoCount
	}
	first := true
	pesLen := len(pesPacket)
	current := 0
	for {
		if current == pesLen {
			break
		}

		var packet []byte
		pcr := first && p.IsVideoKeyFrame()

		if pcr {
			if pesLen-current >= TSPCRPacketLen {
				packet = pesPacket[current : current+TSPCRPacketLen]
				current += TSPCRPacketLen
			} else {
				packet = pesPacket[current:]
				current += pesLen - current

			}
		} else {
			if pesLen-current >= TSDefaultPacketLen {
				packet = pesPacket[current : current+TSDefaultPacketLen]
				current += TSDefaultPacketLen

			} else {
				packet = pesPacket[current:]
				current += pesLen - current

			}
		}
		ts = append(ts, makeTSPacket(packet, pid, counter, first, p.IsVideo, pcr, dts)...)
		if first {
			first = false
		}
		if counter == 0xf {
			counter = 0
		} else {
			counter++
		}
	}
	if p.IsVideo {
		VideoCount = counter
	} else {
		AudioCount = counter
	}

	return ts

}

var VideoCount, AudioCount int = 0, 0

func makeTSPacket(data []byte, pid, counter int, first, isVideo, pcr bool, dts int64) []byte {
	var ts bytes.Buffer
	ts.WriteByte(0x47)
	if first {
		pid |= 0x4000
	}
	ts.WriteByte(byte(pid >> 8))
	ts.WriteByte(byte(pid))

	dataLen := len(data)

	if dataLen != TSDefaultPacketLen {
		counter |= 0x30 //has adaption
	} else {
		counter |= 0x10
	}
	ts.WriteByte(byte(counter))

	if dataLen != TSDefaultPacketLen {
		adaptLen := TSDefaultPacketLen - dataLen - 1
		ts.WriteByte(byte(adaptLen))
		adaptLen -= 1
		if pcr {
			ts.WriteByte(0x50)
			adaptLen -= 6
			ts.Write(addPCR(dts))
		} else {
			if adaptLen >= 0 {
				if first {
					ts.WriteByte(0x40)
				} else {
					ts.WriteByte(0x00)
				}
			}
		}
		for index := 0; index < adaptLen; index++ {
			ts.WriteByte(0xff)
		}
	}
	ts.Write(data)

	return ts.Bytes()
}

/*
	PCR_Ext := 0
	PCR_Const := 0x3F
	PCR_Base := dts
	pcrv := PCR_Ext & 0x1ff
	pcrv |= (PCR_Const << 9) & 0x7E00
	pcrv |= (PCR_Base << 15) & 0xFFFFFFFF8000
	see:
	https://stackoverflow.com/questions/6199940/generate-pcr-from-pts
*/
func addPCR(dts int64) []byte {
	pcr := make([]byte, 6)
	pcr[0] = byte((dts >> 25) & 0xff)

	pcr[1] = byte((dts >> 17) & 0xff)

	pcr[2] = byte((dts >> 9) & 0xff)

	pcr[3] = byte((dts >> 1) & 0xff)

	pcr[4] = byte(((dts & 0x1) << 7) | 0x7e)

	pcr[5] = 0x00
	return pcr
}

func (T *TS) addVideoSeq(data []byte) []byte {
	var video bytes.Buffer
	video.Write(T.VideoSeq)
	video.Write(data)
	return video.Bytes()
}

func addPES(data []byte, isVideo bool, pts, dts int64) []byte {
	var pes bytes.Buffer
	pesStartCode := []byte{0x00, 0x00, 0x01}
	pes.Write(pesStartCode)
	streamID := byte(0xc0)
	if isVideo {
		streamID = 0xe0
	}
	pes.WriteByte(streamID)

	var content bytes.Buffer
	content.WriteByte(0x80)
	flag := byte(0x80)
	pesDataLen := 5
	if isVideo && pts != dts {
		flag |= 0x40
		pesDataLen += 5
	}
	content.WriteByte(flag)
	content.WriteByte(byte(pesDataLen))
	content.Write(makeTimeStamp(flag, pts))
	if isVideo && pts != dts {
		content.Write(makeTimeStamp(0x40, dts))
	}
	if isVideo {
		content.Write([]byte{0x00, 0x00, 0x00, 0x01, 0x09, 0xf0})
	}
	content.Write(data)
	pesPacketLen := content.Len()
	if pesPacketLen > 0xffff {
		pesPacketLen = 0
	}
	pes.WriteByte(byte(pesPacketLen >> 8))
	pes.WriteByte(byte(pesPacketLen))
	pes.Write(content.Bytes())
	return pes.Bytes()
}

/*
	make PTS OR DTS
*/
func makeTimeStamp(flag byte, ts int64) []byte {
	timeStamp := make([]byte, 5)
	if ts > 0x1ffffffff {
		ts -= 0x1ffffffff
	}
	val := int64(0)
	val = (((ts >> 30) & 0x07) << 1) | 1
	timeStamp[0] &= 0x00
	timeStamp[0] |= flag >> 2
	timeStamp[0] |= byte(val)
	val = (((ts >> 15) & 0x7fff) << 1) | 1
	timeStamp[1] = byte(val >> 8)
	timeStamp[2] = byte(val)
	val = ((ts & 0x7fff) << 1) | 1
	timeStamp[3] = byte(val >> 8)
	timeStamp[4] = byte(val)
	return timeStamp
}

func (T *TS) addADTS(data []byte) []byte {
	aacFrameLen := uint16(len(data)) + ADTSLen
	adts := make([]byte, ADTSLen)
	adts[0] = 0xff
	adts[1] = 0xf1

	adts[2] &= 0x00
	adts[2] |= (0x01 << 6)
	adts[2] |= (T.SampleRateIndex << 2)
	adts[2] |= (T.ChannelIndex >> 2)

	adts[3] &= 0x00
	adts[3] |= (T.ChannelIndex << 6)
	adts[3] |= byte(aacFrameLen >> 11)

	adts[4] &= 0x00
	adts[4] |= byte((aacFrameLen >> 3) & 0xFF)

	adts[5] &= 0x00
	adts[5] |= byte((aacFrameLen & 0x07) << 5)
	adts[5] |= byte(0x7f >> 2)
	adts[6] = 0xfc
	var adtsbuf bytes.Buffer
	adtsbuf.Write(adts)
	adtsbuf.Write(data)
	return adtsbuf.Bytes()
}

/*
	Default 44100Hz 2 channels
*/
func (T *TS) getAudioSeqConfigInfo() {
	T.SampleRateIndex = 4
	T.ChannelIndex = 2
	if len(T.AudioSeq) != 0 {
		T.SampleRateIndex = byte((T.AudioSeq[0]&0x07)<<1 | T.AudioSeq[1]>>7)
		T.ChannelIndex = byte(T.AudioSeq[1] >> 3 & 0x0f)
	}

}

var NaluStartCode []byte = []byte{0x00, 0x00, 0x00, 0x01}
var NaluStartCodeLen int = 4

/*
	video seq struct
	0x01+sps[1]+sps[2]+sps[3]+0xFF+0xE1+sps size+sps+01+pps size+pps
*/
func convertVideoSeq(data []byte) []byte {
	videoSeq := make([]byte, 1024)
	start := 6
	spsLen := binary.BigEndian.Uint16(data[start : start+2])
	start += 2
	copy(videoSeq, NaluStartCode)
	sps := videoSeq[NaluStartCodeLen:]
	copy(sps, data[start:start+int(spsLen)])
	start = start + int(spsLen) + 1
	ppsLen := binary.BigEndian.Uint16(data[start : start+2])
	start += 2
	pps := sps[spsLen:]
	copy(pps, NaluStartCode)
	copy(pps[NaluStartCodeLen:], data[start:start+int(ppsLen)])
	return videoSeq[:NaluStartCodeLen*2+int(spsLen+ppsLen)]
}

/*
	change nalu size to nalu start code: 0x00, 0x00, 0x00, 0x01
*/
func convertVideo(data []byte) []byte {
	var current, total int
	total = len(data)

	for current != total {
		header := data[current : current+NaluStartCodeLen]
		size := binary.BigEndian.Uint32(header)
		copy(header, NaluStartCode)
		current += NaluStartCodeLen + int(size)
	}
	return data
}

func MakePAT() []byte {
	var pat bytes.Buffer

	header := []byte{0x47, 0x40, 0x00, 0x10, 0x00}
	pat.Write(header)
	content := []byte{0x00, 0xb0, 0x0d, 0x00, 0x01, 0xc1, 0x00, 0x00, 0x00, 0x01, 0xf0, 0x01, 0x2e, 0x70, 0x19, 0x05}
	pat.Write(content)
	loop := TSPacketLen - pat.Len()
	for index := 0; index < loop; index++ {
		pat.WriteByte(0xFF)
	}
	return pat.Bytes()
}

func (T *TS) MakePMT() []byte {
	var pmt bytes.Buffer
	header := []byte{0x47, 0x50, 0x01, 0x11, 0x00}
	pmt.Write(header)
	sectionLen := []byte{0x02, 0xB0, 0x12}

	if T.hasAudio && T.hasVideo {
		sectionLen = []byte{0x02, 0xB0, 0x17}
	}
	pmt.Write(sectionLen)
	pmt.Write([]byte{0x00, 0x01, 0xC1, 0x00, 0x00, 0xE1, 0x00, 0xF0, 0x00})

	if T.hasVideo {
		pmt.Write([]byte{0x1B, 0xE1, 0x00, 0xF0, 0x00})
	}
	if T.hasAudio {
		pmt.Write([]byte{0x0F, 0xE1, 0x01, 0xF0, 0x00})
	}
	crypt := pmt.Bytes()
	crc32 := GenCRC32(crypt[5:])
	pmt.Write(crc32)
	loop := TSPacketLen - pmt.Len()
	for index := 0; index < loop; index++ {
		pmt.WriteByte(0xFF)
	}
	return pmt.Bytes()
}
