package mp4

import "bytes"

type Tkhd struct {
	Duration uint32
	IsVideo  bool
	Width    uint32
	Height   uint32
}

func (T *Tkhd) Size() uint32 {
	size := BoxHeaderSize + 84
	return uint32(size)
}

func (T *Tkhd) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(T.Size()))
	content.WriteString("tkhd")
	// version 1byte flags 3bytes
	content.Write([]byte{0x00, 0x00, 0x00, 0x03})
	//createTime
	content.Write([]byte{0x00, 0x00, 0x00, 0x00})
	//modificationTime
	content.Write([]byte{0x00, 0x00, 0x00, 0x00})

	// track id 4bytes
	if T.IsVideo {
		content.Write(Mp4Uint32BE(1))
	} else {
		content.Write(Mp4Uint32BE(2))
	}

	// reserved 4bytes
	content.Write(make([]byte, 4))

	content.Write(Mp4Uint32BE(T.Duration))
	// reserved 8bytes
	content.Write(make([]byte, 8))

	if T.IsVideo {
		//layer 2bytes alternate group 2bytes
		content.Write([]byte{0x00, 0x00, 0x00, 0x00})
		content.Write([]byte{0x00, 0x00})
	} else {
		//layer 2bytes alternate group 2bytes
		content.Write([]byte{0x00, 0x00, 0x00, 0x01})
		content.Write([]byte{0x01, 0x00})
	}
	// reserved 4bytes
	content.Write(make([]byte, 2))

	// matrix
	content.Write([]byte{
		0x00, 0x01, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x01, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x40, 0x00, 0x00, 0x00,
	})
	width := Mp4Uint32BE(T.Width)
	content.Write(width[2:])
	content.Write(width[:2])

	height := Mp4Uint32BE(T.Height)
	content.Write(height[2:])
	content.Write(height[:2])

	return content.Bytes()
}
