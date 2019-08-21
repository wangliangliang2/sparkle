package mp4

import "bytes"

type Moov struct {
	Mvhd      Mvhd
	VideoTrak Trak
	AudioTrak Trak
}

func (M *Moov) Size() uint32 {
	size := BoxHeaderSize + M.Mvhd.Size() + M.VideoTrak.Size() + M.AudioTrak.Size()
	return uint32(size)
}

func (M *Moov) Serial() []byte {
	var content bytes.Buffer
	content.Write(Mp4Uint32BE(M.Size()))
	content.WriteString("moov")
	content.Write(M.Mvhd.Serial())
	content.Write(M.VideoTrak.Serial())
	content.Write(M.AudioTrak.Serial())
	return content.Bytes()
}
