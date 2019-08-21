package mp4

import "bytes"

type Smhd struct {
}

func (S *Smhd) Size() uint32 {

	return 16
}

func (S *Smhd) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(S.Size()))
	content.WriteString("smhd")
	// version 1byte flags 3bytes
	content.Write([]byte{0x00, 0x00, 0x00, 0x00})
	// balance 2bytes reserved 2bytes
	content.Write([]byte{0x00, 0x00, 0x00, 0x00})

	return content.Bytes()
}

func DefaultSmhd() Smhd {
	return Smhd{}
}
