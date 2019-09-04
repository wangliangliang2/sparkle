package mp4

import "bytes"

type Mp4a struct {
	Channels   uint32
	SampleSize uint32
	SampleRate uint32
	Esds       Esds
}

func (M *Mp4a) Size() uint32 {

	return BoxHeaderSize + 28 + M.Esds.Size()
}

func (M *Mp4a) Serial() []byte {
	var content bytes.Buffer

	content.Write(Mp4Uint32BE(M.Size()))
	content.WriteString("mp4a")
	//reserved
	content.Write(make([]byte, 6))
	// data_reference_index = 1
	content.Write([]byte{0x00, 0x01})
	//reserved
	content.Write(make([]byte, 8))

	channel := Mp4Uint32BE(M.Channels)
	content.Write(channel[2:])

	sampleSize := Mp4Uint32BE(M.SampleSize)
	content.Write(sampleSize[2:])
	//pre-defined
	content.Write(make([]byte, 2))
	//reserved
	content.Write(make([]byte, 2))

	sampleRate := Mp4Uint32BE(M.SampleRate)
	content.Write(sampleRate[2:])
	content.Write(sampleRate[:2])
	content.Write(M.Esds.Serial())
	return content.Bytes()
}
