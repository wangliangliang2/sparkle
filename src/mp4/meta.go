package mp4

type Meta struct {
}

func (M *Meta) Size() uint32 {

	return uint32(0)
}

func (M *Meta) Serial() []byte {
	return nil
}
