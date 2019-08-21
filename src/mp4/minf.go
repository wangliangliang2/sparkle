package mp4

import "bytes"

type Minf struct {
	Vmhd    Vmhd
	Smhd    Smhd
	Dinf    Dinf
	Stbl    Stbl
	IsVideo bool
}

func (M *Minf) Size() uint32 {
	size := BoxHeaderSize + M.Dinf.Size() + M.Stbl.Size()
	if M.IsVideo {
		size += M.Vmhd.Size()
	} else {
		size += M.Smhd.Size()
	}

	return uint32(size)
}

func (M *Minf) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(M.Size()))
	content.WriteString("minf")
	if M.IsVideo {
		content.Write(M.Vmhd.Serial())
	} else {
		content.Write(M.Smhd.Serial())
	}
	content.Write(M.Dinf.Serial())
	content.Write(M.Stbl.Serial())
	return content.Bytes()
}
