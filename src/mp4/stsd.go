package mp4

import "bytes"

type Stsd struct {
	Avc1    Avc1
	Mp4a    Mp4a
	IsVideo bool
}

func (S *Stsd) Size() uint32 {
	versionAndFlags, numberOfEntry := 4, 4
	size := uint32(BoxHeaderSize + versionAndFlags + numberOfEntry)
	if S.IsVideo {
		size += S.Avc1.Size()
	} else {
		size += S.Mp4a.Size()
	}
	return size
}

func (S *Stsd) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(S.Size()))
	content.WriteString("stsd")
	// version 1byte flags 3bytes
	content.Write([]byte{0x00, 0x00, 0x00, 0x00})
	// number of entiry
	content.Write(Mp4Uint32BE(1))
	if S.IsVideo {
		content.Write(S.Avc1.Serial())
	} else {
		content.Write(S.Mp4a.Serial())
	}
	return content.Bytes()
}
