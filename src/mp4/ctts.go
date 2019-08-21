package mp4

import "bytes"

type Ctts struct {
	EntryCount             uint32
	CompositionOffsetTable []byte
}

func (C *Ctts) Size() uint32 {

	return BoxHeaderSize + 4 + 4 + 8*C.EntryCount
}

func (C *Ctts) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(C.Size()))
	content.WriteString("ctts")
	// version 1byte flags 3bytes
	content.Write([]byte{0x00, 0x00, 0x00, 0x00})
	content.Write(Mp4Uint32BE(C.EntryCount))
	content.Write(C.CompositionOffsetTable)
	return content.Bytes()
}
