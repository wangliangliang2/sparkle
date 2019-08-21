package mp4

import "bytes"

type Stbl struct {
	Stsd    Stsd
	Stts    Stts
	Stss    Stss
	Ctts    Ctts
	Stsc    Stsc
	Stsz    Stsz
	Stco    Stco
	IsVideo bool
}

func (S *Stbl) Size() uint32 {
	size := BoxHeaderSize + S.Stsd.Size() + S.Stts.Size() + S.Stsc.Size() + S.Stsz.Size() + S.Stco.Size()
	if S.IsVideo {
		size += S.Stss.Size() + S.Ctts.Size()
	}

	return size
}

func (S *Stbl) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(S.Size()))
	content.WriteString("stbl")
	content.Write(S.Stsd.Serial())
	content.Write(S.Stts.Serial())
	if S.IsVideo {
		content.Write(S.Stss.Serial())
		content.Write(S.Ctts.Serial())
	}
	content.Write(S.Stsc.Serial())
	content.Write(S.Stsz.Serial())
	content.Write(S.Stco.Serial())

	return content.Bytes()
}
