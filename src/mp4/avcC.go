package mp4

import "bytes"

type AvcC struct {
	VideoSeq []byte
}

func (A *AvcC) Size() uint32 {

	return BoxHeaderSize + uint32(len(A.VideoSeq))
}

func (A *AvcC) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(A.Size()))
	content.WriteString("avcC")
	content.Write(A.VideoSeq)
	return content.Bytes()
}
