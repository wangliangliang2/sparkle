package mp4

import "bytes"

type Vmhd struct {
}

func (V *Vmhd) Size() uint32 {

	return BoxHeaderSize + 12
}

func (V *Vmhd) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(V.Size()))
	content.WriteString("vmhd")
	// version 1byte flags 3bytes
	content.Write([]byte{0x00, 0x00, 0x00, 0x01})

	// graphics mode
	content.Write([]byte{0x00, 0x00, 0x00, 0x00})
	// opcolor
	content.Write([]byte{0x00, 0x00, 0x00, 0x00})
	return content.Bytes()
}

func DefaultVmhd() Vmhd {
	return Vmhd{}
}
