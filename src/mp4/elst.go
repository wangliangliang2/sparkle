package mp4

import "bytes"

type Elst struct {
	NumberOfEntries uint32
	EditListTable   []byte
}

func (E *Elst) Size() uint32 {

	return BoxHeaderSize + 4 + 4 + E.NumberOfEntries*12
}

func (E *Elst) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(E.Size()))
	content.WriteString("elst")
	// version 1byte flags 3bytes
	content.Write([]byte{0x00, 0x00, 0x00, 0x00})

	content.Write(Mp4Uint32BE(E.NumberOfEntries))
	content.Write(E.EditListTable)

	return content.Bytes()
}
