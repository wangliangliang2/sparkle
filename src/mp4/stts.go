package mp4

import "bytes"

type Stts struct {
	NumberOfEntries   uint32
	TimeToSampleTable []byte
}

func (S *Stts) Size() uint32 {

	return BoxHeaderSize + 4 + 4 + 8*S.NumberOfEntries
}

func (S *Stts) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(S.Size()))
	content.WriteString("stts")
	// version 1byte flags 3bytes
	content.Write([]byte{0x00, 0x00, 0x00, 0x00})
	content.Write(Mp4Uint32BE(S.NumberOfEntries))
	content.Write(S.TimeToSampleTable)
	return content.Bytes()
}
