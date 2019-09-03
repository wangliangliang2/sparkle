package mp4

import "bytes"

type Mdat struct {
	Data []byte
}

const largesize uint64 = 8

func (M *Mdat) Size() uint64 {
	if M.isSizeTooBig() {
		return BoxHeaderSize + largesize + uint64(len(M.Data))
	}
	return BoxHeaderSize + uint64(len(M.Data))
}

func (M *Mdat) isSizeTooBig() bool {
	if len(M.Data) > 0xffffffff {
		return true
	}
	return false
}

func (M *Mdat) Serial() []byte {
	var content bytes.Buffer
	if M.isSizeTooBig() {
		content.Write(Mp4Uint32BE(1))
		content.Write(Mp4Uint64BE(M.Size()))
	} else {
		content.Write(Mp4Uint32BE(uint32(M.Size())))
	}
	content.WriteString("mdat")
	content.Write(M.Data)
	return content.Bytes()
}
