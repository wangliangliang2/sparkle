package mp4

import "bytes"

type Trak struct {
	Tkhd Tkhd
	Mdia Mdia
}

func (T *Trak) Size() uint32 {
	size := BoxHeaderSize + T.Tkhd.Size() + T.Mdia.Size()
	return uint32(size)
}

func (T *Trak) Serial() []byte {
	var content bytes.Buffer
	content.Write(Mp4Uint32BE(T.Size()))
	content.WriteString("trak")
	content.Write(T.Tkhd.Serial())
	content.Write(T.Mdia.Serial())
	return content.Bytes()
}
