package mp4

import "bytes"

type Edts struct {
	Elst Elst
}

func (E *Edts) Size() uint32 {

	return BoxHeaderSize + 28
}

func (E *Edts) Serial() []byte {
	var content bytes.Buffer
	content.Write(Mp4Uint32BE(E.Size()))
	content.WriteString("edts")
	content.Write(E.Elst.Serial())
	return content.Bytes()
}
