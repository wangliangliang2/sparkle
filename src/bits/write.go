package bits

func PutU16BE(v uint16) (b []byte) {
	b = PutUBE(uint64(v), 2)
	return
}

func PutU24BE(v uint32) (b []byte) {
	b = PutUBE(uint64(v), 3)
	return
}

func PutU32BE(v uint32) (b []byte) {
	b = PutUBE(uint64(v), 4)
	return
}

func PutU64BE(v uint64) (b []byte) {
	b = PutUBE(v, 8)
	return
}

func PutUBE(v uint64, n int) (b []byte) {
	b = make([]byte, n)
	size := n - 1
	for index := size * 8; index >= 0; index -= 8 {
		b[size-index/8] = byte(v >> uint64(index))
	}
	return
}

func PutU32LE(v uint32) (b []byte) {
	b = PutULE(uint64(v), 4)
	return
}

func PutULE(v uint64, n int) (b []byte) {
	b = make([]byte, n)
	size := n - 1
	for index := size * 8; index >= 0; index -= 8 {
		b[index/8] = byte(v >> uint64(index))
	}
	return

}
