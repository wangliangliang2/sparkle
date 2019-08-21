package mp4

import "bytes"

type Hdlr struct {
	Name    string
	IsVideo bool
}

func (H *Hdlr) Size() uint32 {
	size := BoxHeaderSize + 24 + len(H.Name) + 1
	return uint32(size)
}

func (H *Hdlr) Serial() []byte {
	var content bytes.Buffer
	content.Write(Mp4Uint32BE(H.Size()))
	content.WriteString("hdlr")
	// version 1byte flags 3bytes
	content.Write([]byte{0x00, 0x00, 0x00, 0x00})
	//pre-defined
	content.Write([]byte{0x00, 0x00, 0x00, 0x00})
	if H.IsVideo {
		content.Write([]byte{0x76, 0x69, 0x64, 0x65})
	} else {
		content.Write([]byte{0x73, 0x6F, 0x75, 0x6E})
	}
	content.Write(make([]byte, 12))

	content.WriteString(H.Name)
	content.WriteByte(0)
	return content.Bytes()
}
