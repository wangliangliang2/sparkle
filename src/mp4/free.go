package mp4

type Free struct {
}

func (F *Free) Size() uint32 {
	return 8
}

func (F *Free) Serial() []byte {
	return []byte{0x00, 0x00, 0x00, 0x08, 0x66, 0x72, 0x65, 0x65}
}
