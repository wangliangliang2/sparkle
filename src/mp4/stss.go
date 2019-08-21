package mp4

import "bytes"

type Stss struct {
	NumberOfEntries uint32
	SyncSampleTable []byte
}

func (S *Stss) Size() uint32 {

	return BoxHeaderSize + 4 + 4 + 4*S.NumberOfEntries
}

func (S *Stss) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(S.Size()))
	content.WriteString("stss")
	// version 1byte flags 3bytes
	content.Write([]byte{0x00, 0x00, 0x00, 0x00})
	content.Write(Mp4Uint32BE(S.NumberOfEntries))
	content.Write(S.SyncSampleTable)
	return content.Bytes()
}
