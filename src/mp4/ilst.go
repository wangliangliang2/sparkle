package mp4

type Ilst struct {
}

func (I *Ilst) Size() uint32 {

	return uint32(0)
}

func (I *Ilst) Serial() []byte {
	return nil
}
