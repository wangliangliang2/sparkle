package mp4

import "bytes"

type Mdhd struct {
	TimeScale uint32
	Duration  uint32
}

func (M *Mdhd) Size() uint32 {

	return BoxHeaderSize + 24
}

func (M *Mdhd) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(M.Size()))
	content.WriteString("mdhd")
	// version 1byte flags 3bytes
	content.Write([]byte{0x00, 0x00, 0x00, 0x00})
	//createTime
	content.Write([]byte{0x00, 0x00, 0x00, 0x00})
	//modificationTime
	content.Write([]byte{0x00, 0x00, 0x00, 0x00})

	content.Write(Mp4Uint32BE(M.TimeScale))
	content.Write(Mp4Uint32BE(M.Duration))
	// language: und (undetermined)  pre_defined = 0
	content.Write([]byte{0x55, 0xC4, 0x00, 0x00})
	return content.Bytes()
}
