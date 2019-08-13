package client

import (
	"av"
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"protocol/flv/tag"
	"time"
)

type Flv struct {
	isOld    bool
	isClosed bool
	uri      string
	token    string
	Notice   chan struct{}
	handle   http.ResponseWriter
}

func NewFlvClient(uri string, w http.ResponseWriter) *Flv {
	flv := &Flv{
		uri:    uri,
		Notice: make(chan struct{}),
		handle: w,
	}
	flv.Token()
	return flv
}

func (F Flv) IsHls() (ret bool) {
	ret = false
	return
}

func (F *Flv) IsOld() (ret bool) {
	ret = F.isOld
	F.isOld = true
	return
}

func (F Flv) Uri() (ret string) {
	ret = F.uri
	return
}

func (F *Flv) LeaveShow() {
	F.Notice <- struct{}{}
}

func (F *Flv) CloseChannel() {
	close(F.Notice)
}

func (F *Flv) Token() (ret string) {
	if F.token != "" {
		ret = F.token
		return
	}
	text := F.uri + time.Now().Format("2006-01-02 15:04:05")
	hasher := md5.New()
	hasher.Write([]byte(text))
	F.token = hex.EncodeToString(hasher.Sum(nil))
	ret = F.token
	return
}

func (F *Flv) ReceivePacket(packet *av.Packet) {
	F.Write(packet.ToFlvTag())
}

func (F *Flv) Work() {
	F.Write(tag.FlvFileHeader)
	for {
		time.Sleep(time.Second)
		if F.isClosed {
			F.LeaveShow()
			break
		}
	}
	F.CloseChannel()
}

func (F *Flv) Write(b []byte) {
	var err error
	if _, err = F.handle.Write(b); err != nil {
		F.isClosed = true
		return
	}
}

func (F *Flv) Close() {
	F.isClosed = true
}

func (F Flv) GetPacket() (packet *av.Packet, ok bool) {
	return
}
