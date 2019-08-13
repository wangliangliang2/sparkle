package buff

import (
	"bufio"
	"errors"
	"io"
	"net"
	"time"
)

const BufSize = 250000

var (
	BytesError   = errors.New("n must <= 4")
	TimeoutError = errors.New("Read Timeout...")
)

var IOTimeout = 10 * time.Second

type ReadWriter struct {
	conn net.Conn
	*bufio.ReadWriter
}

func New(rw net.Conn) *ReadWriter {
	return &ReadWriter{
		rw,
		bufio.NewReadWriter(
			bufio.NewReaderSize(rw, BufSize),
			bufio.NewWriterSize(rw, BufSize),
		),
	}
}

func (RW *ReadWriter) Read(p []byte) (n int, err error) {
	RW.conn.SetReadDeadline(time.Now().Add(IOTimeout))
	n, err = io.ReadFull(RW.ReadWriter, p)
	RW.conn.SetReadDeadline(time.Time{})
	return
}

func (RW *ReadWriter) OneByte() (b byte, err error) {
	RW.conn.SetReadDeadline(time.Now().Add(IOTimeout))
	b, err = RW.ReadByte()
	RW.conn.SetReadDeadline(time.Time{})
	return
}

func (RW *ReadWriter) Peeks(n int) (b []byte, err error) {
	RW.conn.SetReadDeadline(time.Now().Add(IOTimeout))
	b, err = RW.Peek(n)
	RW.conn.SetReadDeadline(time.Time{})
	return
}

func (RW *ReadWriter) WriteAtOnce(p []byte) (n int, err error) {
	n, err = RW.Write(p)
	if err != nil {
		return
	}
	err = RW.Flush()
	return
}

func (RW *ReadWriter) ReadUint32BE(n int) (ret uint32, err error) {
	if err = byteErr(n); err != nil {
		return
	}
	var b byte
	for index := 0; index < n; index++ {
		if b, err = RW.OneByte(); err != nil {
			return
		}
		ret = ret<<8 + uint32(b)
	}
	return
}

func (RW *ReadWriter) ReadUint32LE(n int) (ret uint32, err error) {
	if err = byteErr(n); err != nil {
		return
	}
	var b byte
	for index := 0; index < n; index++ {
		if b, err = RW.OneByte(); err != nil {
			return
		}
		ret += uint32(b) << uint32(index*8)
	}
	return
}

func (RW *ReadWriter) WriteUint32BE(val uint32, n int) (err error) {
	if err = byteErr(n); err != nil {
		return
	}
	b := make([]byte, n)
	for index := 0; index < n; index++ {
		b[index] = byte(val >> uint((n-index-1)*8))
	}
	_, err = RW.Write(b)
	return
}

func (RW *ReadWriter) WriteUint32LE(val uint32, n int) (err error) {
	if err = byteErr(n); err != nil {
		return
	}
	b := make([]byte, n)
	for index := 0; index < n; index++ {
		b[index] = byte(val >> uint(index*8))
	}
	_, err = RW.Write(b)
	return
}

func byteErr(n int) (err error) {
	if n > 4 {
		err = BytesError
		return
	}
	return
}
