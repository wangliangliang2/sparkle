package center

import (
	"av"
	"bytes"
	"mp4"
	"os"
	"time"
)

type Mp4 struct {
	VideoSampleCount     uint32
	StartTime            time.Time
	VideoStssBuffer      bytes.Buffer
	VideoSeq             []byte
	LastVideoCttsOffet   uint32
	VideoCttsBuffer      bytes.Buffer
	VideoCttsRepeatCount uint32
	VideoStszBuffer      bytes.Buffer
	VideoStcoBuffer      bytes.Buffer
	VideoSttsBuffer      bytes.Buffer
	VideoStsc            []byte
	Offset               uint32
	Data                 bytes.Buffer

	AudioSampleCount uint32
	AudioStcoBuffer  bytes.Buffer
	AudioStszBuffer  bytes.Buffer
	AudioSttsBuffer  bytes.Buffer
	AudioStsc        []byte
	AudioSeq         []byte
	AacInfo          mp4.AACInfo
	FtypBox          mp4.Ftyp
	FreeBox          mp4.Free
	Duration         uint32
	VideoWidth       uint32
	VideoHeight      uint32
}

var Stsc []byte = []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01}

func NewMp4() (ret *Mp4) {
	ret = &Mp4{
		StartTime: time.Now(),
		FtypBox:   mp4.GetFtypBox(),
		FreeBox:   mp4.GetFreeBox(),
		VideoStsc: Stsc,
		AudioStsc: Stsc,
	}
	ret.Offset = uint32(len(ret.FtypBox.Serial())) + uint32(len(ret.FreeBox.Serial())) + 8
	return
}

func (M *Mp4) StartStore(packet *av.Packet) {
	if packet.IsVideoSequence() {
		M.VideoSeq = packet.Data[5:]
	}
	if packet.IsAudioSequence() {
		M.AacInfo, _ = mp4.AudioAACInfo(packet.Data)
		M.AudioSeq = packet.Data[2:]
	}
	if packet.IsPureVideoData() {
		M.countVideoSample()
		M.storeVideoStco(packet)
		M.storeVideoStsz(packet)
		M.storeVideoStss(packet)
		M.storeVideoCtts(packet)
		M.storeVideoData(packet)
	}
	if packet.IsPureAudioData() {
		M.countAudioSample()
		M.storeAudioStco(packet)
		M.storeAudioStsz(packet)
		M.storeAudioData(packet)
	}
}

func (M *Mp4) countVideoSample() {
	M.VideoSampleCount++
}

func (M *Mp4) countAudioSample() {
	M.AudioSampleCount++
}

func (M *Mp4) StopStore(filename string) {
	M.storeStts()
	moov, mdat := M.GetMp4MoovAndMdatBox()
	WriteTofile(filename, M.FtypBox.Serial(), M.FreeBox.Serial(), mdat.Serial(), moov.Serial())
}

func (M *Mp4) GetMp4MoovAndMdatBox() (moov mp4.Moov, mdat mp4.Mdat) {
	M.VideoWidth, M.VideoHeight = av.GetSizeFromVideoSeq(M.VideoSeq)
	moov = mp4.Moov{
		Mvhd:      mp4.GetMvhdBox(M.Duration),
		VideoTrak: mp4.GetVideoTrak(M.Duration, M.VideoWidth, M.VideoHeight, M.VideoSeq, M.VideoSttsBuffer.Bytes(), M.VideoStssBuffer.Bytes(), M.VideoCttsBuffer.Bytes(), M.VideoStsc, M.VideoStszBuffer.Bytes(), M.VideoStcoBuffer.Bytes()),
		AudioTrak: mp4.GetAudioTrak(M.AacInfo, M.Duration, M.AudioSeq, M.AudioSttsBuffer.Bytes(), M.AudioStsc, M.AudioStszBuffer.Bytes(), M.AudioStcoBuffer.Bytes()),
	}
	mdat = mp4.Mdat{
		Data: M.Data.Bytes(),
	}
	return
}

func WriteTofile(filename string, ftyp, free, mdat, moov []byte) {
	dir, _ := os.Getwd()
	f, _ := os.Create(dir + "/" + filename + ".mp4")
	defer f.Close()
	f.Write(ftyp)
	f.Write(free)
	f.Write(mdat)
	f.Write(moov)
}

func (M *Mp4) storeStts() {
	endTime := time.Now()
	M.Duration = uint32(endTime.Sub(M.StartTime).Seconds())
	M.storeVideoStts(M.Duration)
	M.storeAudioStts(M.Duration)
}

func (M *Mp4) storeAudioStco(packet *av.Packet) {
	M.AudioStcoBuffer.Write(mp4.Mp4Uint32BE(M.Offset))
	M.Offset += uint32(len(packet.Data[2:]))
}

func (M *Mp4) storeAudioStsz(packet *av.Packet) {
	M.AudioStszBuffer.Write(mp4.Mp4Uint32BE(uint32(len(packet.Data[2:]))))
}

func (M *Mp4) storeAudioStts(duration uint32) {
	aacInfo := M.AacInfo
	audioSampleDuration := duration * aacInfo.SampleRate / M.AudioSampleCount
	M.AudioSttsBuffer.Write(mp4.Mp4Uint32BE(M.AudioSampleCount))
	M.AudioSttsBuffer.Write(mp4.Mp4Uint32BE(audioSampleDuration))
}

func (M *Mp4) storeAudioData(packet *av.Packet) {
	M.Data.Write(packet.Data[2:])
}

func (M *Mp4) storeVideoStco(packet *av.Packet) {
	M.VideoStcoBuffer.Write(mp4.Mp4Uint32BE(M.Offset))
	M.Offset += uint32(len(packet.Data[5:]))
}

func (M *Mp4) storeVideoStsz(packet *av.Packet) {
	M.VideoStszBuffer.Write(mp4.Mp4Uint32BE(uint32(len(packet.Data[5:]))))
}

func (M *Mp4) storeVideoStss(packet *av.Packet) {
	if packet.IsVideoKeyFrame() {
		M.VideoStssBuffer.Write(mp4.Mp4Uint32BE(M.VideoSampleCount))
	}
}

func (M *Mp4) storeVideoCtts(packet *av.Packet) {
	if packet.CompositionTime != M.LastVideoCttsOffet {
		M.VideoCttsBuffer.Write(mp4.Mp4Uint32BE(M.VideoCttsRepeatCount))
		M.VideoCttsBuffer.Write(mp4.Mp4Uint32BE(packet.CompositionTime))
		M.VideoCttsRepeatCount = 0
		M.LastVideoCttsOffet = packet.CompositionTime
	}
	M.VideoCttsRepeatCount++
}

func (M *Mp4) storeVideoStts(duration uint32) {
	videoSampleDuration := duration * 1000 * 30 / M.VideoSampleCount
	M.VideoSttsBuffer.Write(mp4.Mp4Uint32BE(M.VideoSampleCount))
	M.VideoSttsBuffer.Write(mp4.Mp4Uint32BE(videoSampleDuration))
}

func (M *Mp4) storeVideoData(packet *av.Packet) {
	M.Data.Write(packet.Data[5:])
}
