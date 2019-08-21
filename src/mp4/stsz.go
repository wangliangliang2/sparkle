package mp4

import "bytes"

type Stsz struct {
	SampleSize      uint32
	NumberOfEntries uint32
	SampleSizeTable []byte
}

func (S *Stsz) Size() uint32 {

	return BoxHeaderSize + 4 + 4 + 4 + S.NumberOfEntries*4
}

func (S *Stsz) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(S.Size()))
	content.WriteString("stsz")
	// version 1byte flags 3bytes
	content.Write([]byte{0x00, 0x00, 0x00, 0x00})
	content.Write(Mp4Uint32BE(S.SampleSize))
	content.Write(Mp4Uint32BE(S.NumberOfEntries))
	content.Write(S.SampleSizeTable)
	return content.Bytes()
}
