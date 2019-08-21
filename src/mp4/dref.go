package mp4

import "bytes"

// data 00 00 00 0C 75 72 6C 20 00 00 00 01
type Dref struct {
	NumberOfEntries uint32
	Data            []byte
}

func (D *Dref) Size() uint32 {
	size := BoxHeaderSize + 4 + 4 + 12*D.NumberOfEntries
	return uint32(size)
}

func (D *Dref) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(D.Size()))
	content.WriteString("dref")

	// version 1byte flags 3bytes
	content.Write([]byte{0x00, 0x00, 0x00, 0x00})
	content.Write(Mp4Uint32BE(D.NumberOfEntries))
	content.Write(D.Data)
	return content.Bytes()
}

func DefaultDref() Dref {
	return Dref{
		NumberOfEntries: 1,
		Data:            []byte{0x00, 0x00, 0x00, 0x0C, 0x75, 0x72, 0x6C, 0x20, 0x00, 0x00, 0x00, 0x01},
	}
}
