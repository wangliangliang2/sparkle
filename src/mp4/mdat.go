package mp4

import "bytes"

type Mdat struct {
	Data []byte
}

func (M *Mdat) Size() uint32 {

	return BoxHeaderSize + uint32(len(M.Data))
}

func (M *Mdat) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(M.Size()))
	content.WriteString("mdat")
	content.Write(M.Data)
	return content.Bytes()
}
