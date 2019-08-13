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
	NaluStartCodeLen   = 4
)

var (
	NaluStartCode = []byte{0x00, 0x00, 0x00, 0x01}
	pesStartCode  = []byte{0x00, 0x00, 0x01}
)

type MPEG4AudioCondig struct {
	SampleRateIndex byte
	ChannelIndex    byte
}

type TS struct {
	hasVideo, hasAudio bool
	AudioSeq, VideoSeq []byte
	MPEG4AudioCondig
	VideoCount, AudioCount int
}

func GetPAT() (ret []byte) {
	header := []byte{0x47, 0x40, 0x00, 0x10, 0x00}
	playload := []byte{0x00, 0xb0, 0x0d, 0x00, 0x01, 0xc1, 0x00, 0x00, 0x00, 0x01, 0xf0, 0x01, 0x2e, 0x70, 0x19, 0x05}
	var pat bytes.Buffer
	pat.Write(header)
	pat.Write(playload)
	stuffing := TSPacketLen - pat.Len()
	pat.Write(fillStuffing(stuffing))
	return pat.Bytes()
}

func fillStuffing(stuffing int) (ret []byte) {
	var stuff bytes.Buffer
	for index := 0; index < stuffing; index++ {
		stuff.WriteByte(0xFF)
	}
	ret = stuff.Bytes()
	return
}

func (T *TS) GetPMT() []byte {
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
	stuffing := TSPacketLen - pmt.Len()
	pmt.Write(fillStuffing(stuffing))
	return pmt.Bytes()
}

func (T *TS) ReadSeq(p *av.Packet) {
	switch {
	case p.IsAudioSequence():
		T.hasAudio = true
		T.AudioSeq = p.Data[2:] //move flv audio header
		T.setAudioSeqConfigInfo()
	case p.IsVideoSequence():
		T.hasVideo = true
		data := p.Data[5:] // move flv video header
		T.VideoSeq = setVideoSeq(data)
	}
}

/*
	Default 44100Hz 2 channels
*/
func (T *TS) setAudioSeqConfigInfo() {
	T.SampleRateIndex = 4
	T.ChannelIndex = 2
	if len(T.AudioSeq) != 0 {
		T.SampleRateIndex = byte((T.AudioSeq[0]&0x07)<<1 | T.AudioSeq[1]>>7)
		T.ChannelIndex = byte(T.AudioSeq[1] >> 3 & 0x0f)
	}
}

/*
	video seq struct
	0x01+sps[1]+sps[2]+sps[3]+0xFF+0xE1+sps size+sps+01+pps size+pps
*/
func setVideoSeq(data []byte) (ret []byte) {
	spsSize := binary.BigEndian.Uint16(data[6:8])
	ppsSize := binary.BigEndian.Uint16(data[8+spsSize+1 : 8+spsSize+1+2])
	sps := data[8 : 8+spsSize]
	pps := data[8+spsSize+1+2 : 8+spsSize+1+2+ppsSize]

	var videoseq bytes.Buffer
	videoseq.Write(NaluStartCode)
	videoseq.Write(sps)
	videoseq.Write(NaluStartCode)
	videoseq.Write(pps)
	ret = videoseq.Bytes()
	return
}

func (T *TS) ReadData(p *av.Packet) []byte {
	var ts bytes.Buffer
	var payload []byte
	pesPacket, dts, pid, counter := T.getPesPacket(p)
	current, pesLen, first := 0, len(pesPacket), true
	for {
		if current == pesLen {
			break
		}
		addPcr := first && p.IsVideoKeyFrame()
		payload, current = getOneTsPayload(addPcr, pesLen, current, pesPacket)
		ts.Write(TsLayer(payload, pid, counter, first, p.IsVideo, addPcr, dts))
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
		T.VideoCount = counter
	} else {
		T.AudioCount = counter
	}
	return ts.Bytes()

}

func (T *TS) getPesPacket(p *av.Packet) (pesPacket []byte, dts int64, pid, counter int) {
	data := make([]byte, len(p.Data))
	copy(data, p.Data)
	dts = int64(p.Timestamp) * H264DefaultHZ
	switch {
	case p.IsPureAudioData():
		pesPacket, pid, counter = T.getPesPacketFromAudio(data, dts)
	case p.IsPureVideoData():
		pesPacket, pid, counter = T.getPesPacketFromVideo(data, dts, p.CompositionTime, p.IsVideoKeyFrame())
	}
	return
}

func (T *TS) getPesPacketFromAudio(data []byte, dts int64) (pesPacket []byte, pid, counter int) {
	data = data[2:]
	esAudio := T.addADTS(data)
	pesPacket = PesLayer(esAudio, false, dts, 0)
	pid = AudioPID
	counter = T.AudioCount
	return
}

func (T *TS) getPesPacketFromVideo(data []byte, dts int64, compositionTime uint32, isVideoKeyFrame bool) (pesPacket []byte, pid, counter int) {
	data = data[5:]
	esVideo := convertVideo(data)
	if isVideoKeyFrame {
		esVideo = T.addVideoSeq(esVideo)
	}
	pts := dts + int64(compositionTime)*H264DefaultHZ
	pesPacket = PesLayer(esVideo, true, pts, dts)
	pid = VideoPID
	counter = T.VideoCount
	return
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

func (T *TS) addVideoSeq(data []byte) []byte {
	var video bytes.Buffer
	video.Write(T.VideoSeq)
	video.Write(data)
	return video.Bytes()
}

func PesLayer(data []byte, isVideo bool, pts, dts int64) []byte {
	var pes bytes.Buffer
	pes.Write(pesStartCode)
	pes.WriteByte(getPesStreamId(isVideo))
	length, content := getPesContent(data, isVideo, pts, dts)
	pes.Write(getPesLength(length))
	pes.Write(content)
	return pes.Bytes()
}

func getPesStreamId(isVideo bool) (streamId byte) {
	streamId = 0xc0
	if isVideo {
		streamId = 0xe0
	}
	return
}

func getPesContent(data []byte, isVideo bool, pts, dts int64) (length int, ret []byte) {
	var content bytes.Buffer
	content.WriteByte(getPesDataFlag())
	flag, pesDataLen := getPesTimeFlag(isVideo, pts, dts)
	content.WriteByte(flag)
	content.WriteByte(pesDataLen)
	content.Write(getPesPts(flag, pts))
	if isVideo && pts != dts {
		content.Write(getPesDts(0x40, dts))
	}
	if isVideo {
		content.Write([]byte{0x00, 0x00, 0x00, 0x01, 0x09, 0xf0})
	}
	content.Write(data)
	ret = content.Bytes()
	length = content.Len()
	return
}

func getPesTimeFlag(isVideo bool, pts, dts int64) (flag, pesDataLen byte) {
	flag, pesDataLen = 0x80, 5
	if isVideo && pts != dts {
		flag |= 0x40
		pesDataLen += 5
	}
	return
}

func getPesDataFlag() (flag byte) {
	flag = 0x80
	return
}

func getPesPts(flag byte, pts int64) (ret []byte) {
	ret = makeTimeStamp(flag, pts)
	return
}

func getPesDts(flag byte, dts int64) (ret []byte) {
	ret = makeTimeStamp(flag, dts)
	return
}

func getPesLength(length int) (ret []byte) {
	ret = make([]byte, 2)
	if length > 0xffff {
		length = 0
	}
	ret[0] = byte(length >> 8)
	ret[1] = byte(length)
	return
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

func getOneTsPayload(addPcr bool, pesLen, current int, pesPacket []byte) (payload []byte, newPosition int) {
	if addPcr {
		if pesLen-current >= TSPCRPacketLen {
			payload = pesPacket[current : current+TSPCRPacketLen]
			current += TSPCRPacketLen
		} else {
			payload = pesPacket[current:]
			current += pesLen - current

		}
	} else {
		if pesLen-current >= TSDefaultPacketLen {
			payload = pesPacket[current : current+TSDefaultPacketLen]
			current += TSDefaultPacketLen

		} else {
			payload = pesPacket[current:]
			current += pesLen - current

		}
	}
	newPosition = current
	return
}

func TsLayer(data []byte, pid, counter int, first, isVideo, pcr bool, dts int64) []byte {
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
