package mp4

import "bytes"

type Stco struct {
	NumberOfEntries  uint32
	ChunkOffsetTable []byte
}

func (S *Stco) Size() uint32 {

	return BoxHeaderSize + 4 + 4 + S.NumberOfEntries*4
}

func (S *Stco) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(S.Size()))
	content.WriteString("stco")
	// version 1byte flags 3bytes
	content.Write([]byte{0x00, 0x00, 0x00, 0x00})
	content.Write(Mp4Uint32BE(S.NumberOfEntries))
	content.Write(S.ChunkOffsetTable)
	return content.Bytes()
}
