package mp4

import "bytes"

type Avc1 struct {
	Width  uint32
	Height uint32
	AvcC   AvcC
}

func (A *Avc1) Size() uint32 {
	size := BoxHeaderSize + 78 + A.AvcC.Size()
	return size
}

func (A *Avc1) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(A.Size()))
	content.WriteString("avc1")
	// reserved
	content.Write(make([]byte, 6))
	// data_reference_index
	content.Write([]byte{0x00, 0x01})
	// pre_defined
	content.Write(make([]byte, 2))
	// reserved
	content.Write(make([]byte, 2))
	// pre_defined
	content.Write(make([]byte, 12))
	width := Mp4Uint32BE(A.Width)
	height := Mp4Uint32BE(A.Height)
	content.Write(width[2:])
	content.Write(height[2:])
	// horizresolution
	content.Write([]byte{0x00, 0x48, 0x00, 0x00})
	// vertresolution
	content.Write([]byte{0x00, 0x48, 0x00, 0x00})
	// reserved
	content.Write(make([]byte, 4))

	//  frame_count
	content.Write([]byte{0x00, 0x01})

	// compressorname
	content.Write(make([]byte, 32))

	// depth
	content.Write([]byte{0x00, 0x18})

	// pre_defined = -1
	content.Write([]byte{0xFF, 0xFF})

	content.Write(A.AvcC.Serial())
	return content.Bytes()
}
