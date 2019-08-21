package mp4

import "bytes"

type Stsc struct {
	NumberOfEntries    uint32
	SampleToChunkTable []byte
}

func (S *Stsc) Size() uint32 {

	return BoxHeaderSize + 4 + 4 + S.NumberOfEntries*12
}

func (S *Stsc) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(S.Size()))
	content.WriteString("stsc")
	// version 1byte flags 3bytes
	content.Write([]byte{0x00, 0x00, 0x00, 0x00})
	content.Write(Mp4Uint32BE(S.NumberOfEntries))
	content.Write(S.SampleToChunkTable)
	return content.Bytes()

}
