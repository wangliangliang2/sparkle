package mp4

import "bytes"

type Esds struct {
	Audioseq []byte
}

func (E *Esds) Size() uint32 {

	return BoxHeaderSize + 4 + 5 + 7 + 8 + 2 + 3 + uint32(len(E.Audioseq))
}

func (E *Esds) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(E.Size()))
	content.WriteString("esds")
	// version 1byte flags 3bytes
	content.Write([]byte{0x00, 0x00, 0x00, 0x00})

	content.Write([]byte{0x03, 0x1E, 0x00, 0x02, 0x00})

	content.Write([]byte{0x04, 0x16, 0x40, 0x15, 0x00, 0x00, 0x00})

	//bitrate
	content.Write(Mp4Uint32BE(0))
	//bitrate
	content.Write(Mp4Uint32BE(0))
	content.Write([]byte{0x05, byte(len(E.Audioseq))})
	content.Write(E.Audioseq)
	content.Write([]byte{0x06, 0x01, 0x02})

	return content.Bytes()

}
