package mp4

import "bytes"

type Dinf struct {
	Dref Dref
}

func (D *Dinf) Size() uint32 {

	return BoxHeaderSize + D.Dref.Size()
}

func (D *Dinf) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(D.Size()))
	content.WriteString("dinf")
	content.Write(D.Dref.Serial())
	return content.Bytes()
}
