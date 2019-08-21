package mp4

import (
	"bytes"
)

type Ftyp struct {
	MajorBrand       string
	MinorVersion     uint32
	CompatibleBrands string
}

func (F *Ftyp) Size() uint32 {
	remainder := len(F.CompatibleBrands) % 4
	size := BoxHeaderSize + len(F.MajorBrand) + 4 + len(F.CompatibleBrands)
	if remainder != 0 {
		size += 4 - remainder
	}
	return uint32(size)
}

func (F *Ftyp) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(F.Size()))
	content.WriteString("ftyp")
	content.WriteString(F.MajorBrand)
	content.Write(Mp4Uint32BE(F.MinorVersion))
	content.WriteString(F.CompatibleBrands)
	remainder := len(F.CompatibleBrands) % 4
	if remainder != 0 {
		fill := make([]byte, 4-remainder)
		content.Write(fill)
	}
	return content.Bytes()
}
