package mp4

import "bytes"

type Mdia struct {
	Mdhd Mdhd
	Hdlr Hdlr
	Minf Minf
}

func (M *Mdia) Size() uint32 {
	size := BoxHeaderSize + M.Mdhd.Size() + M.Hdlr.Size() + M.Minf.Size()
	return uint32(size)
}

func (M *Mdia) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(M.Size()))
	content.WriteString("mdia")
	content.Write(M.Mdhd.Serial())
	content.Write(M.Hdlr.Serial())
	content.Write(M.Minf.Serial())
	return content.Bytes()
}
