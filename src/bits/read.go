package bits

func U16BE(b []byte) (i uint16) {
	i = uint16(UBE(b, 2))
	return
}

func U24BE(b []byte) (i uint32) {
	i = uint32(UBE(b, 3))
	return
}

func U32BE(b []byte) (i uint32) {
	i = uint32(UBE(b, 4))
	return
}

func U64BE(b []byte) (i uint64) {
	i = UBE(b, 8)
	return
}

func UBE(b []byte, n int) (i uint64) {
	for index := 0; index < n; index++ {
		if index == 0 {
			i = uint64(b[index])
		} else {
			i <<= 8
			i |= uint64(b[index])
		}
	}
	return
}

func U32LE(b []byte) (i uint32) {
	i = uint32(ULE(b, 4))
	return
}

func ULE(b []byte, n int) (i uint64) {
	lastIndex := n - 1
	for index := lastIndex; index >= 0; index-- {
		if index == lastIndex {
			i = uint64(b[index])
		} else {
			i <<= 8
			i |= uint64(b[index])
		}
	}
	return
}
