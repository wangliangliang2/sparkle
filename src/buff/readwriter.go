package buff

import (
	"bufio"
	"io"
)

const (
	BigEndian = iota
	LittleEndian
)

const BufSize = 4 * 1024

type ReadWriter struct {
	*bufio.ReadWriter
	readError  error
	writeError error
}

func New(c io.ReadWriter) *ReadWriter {

	return &ReadWriter{
		ReadWriter: bufio.NewReadWriter(
			bufio.NewReaderSize(c, BufSize),
			bufio.NewWriterSize(c, BufSize),
		),
	}
}

func (RW *ReadWriter) Read(b []byte) (n int, err error) {
	if RW.readError != nil {
		n, err = 0, RW.readError
		return
	}

	n, err = io.ReadAtLeast(RW.ReadWriter, b, len(b))
	RW.readError = err
	return
}

func (RW *ReadWriter) ReadUint32BE(n int) (ret uint32, err error) {
	return RW.ReadUint32(n, BigEndian)
}
func (RW *ReadWriter) ReadUint32LE(n int) (ret uint32, err error) {
	return RW.ReadUint32(n, LittleEndian)
}
func (RW *ReadWriter) ReadUint32(n int, endian int) (ret uint32, err error) {
	if RW.readError != nil {
		ret, err = 0, RW.readError
		return
	}
	for i := 0; i < n; i++ {
		b, tErr := RW.ReadByte()
		if tErr != nil {
			RW.readError = tErr
			ret, err = 0, RW.readError
			return
		}
		if endian == BigEndian {
			ret = ret<<8 + uint32(b)
		} else {
			ret += uint32(b) << uint32(i*8)
		}

	}
	return
}

func (RW *ReadWriter) Write(b []byte) (n int, err error) {
	if RW.writeError != nil {
		n, err = 0, RW.writeError
		return
	}
	n, err = RW.ReadWriter.Write(b)
	return
}

func (RW *ReadWriter) WriteUint32LE(v uint32, n int) (err error) {

	return RW.WriteUint32(v, n, LittleEndian)
}

func (RW *ReadWriter) WriteUint32BE(v uint32, n int) (err error) {

	return RW.WriteUint32(v, n, BigEndian)
}

func (RW *ReadWriter) WriteUint32(v uint32, n int, endian int) (err error) {
	if RW.writeError != nil {
		err = RW.writeError
		return
	}
	for i := 0; i < n; i++ {
		var b byte
		if endian == BigEndian {
			b = byte(v>>uint32((n-i-1)<<3)) & 0xff
		} else {
			b = byte(v) & 0xff
		}
		if err = RW.WriteByte(b); err != nil {
			RW.writeError = err
			return
		}
		if endian == LittleEndian {
			v = v >> 8
		}
	}
	return
}

func (RW *ReadWriter) Flush() error {
	if RW.writeError != nil {
		return RW.writeError
	}
	if RW.ReadWriter.Writer.Buffered() == 0 {
		return nil
	}
	return RW.ReadWriter.Flush()
}
