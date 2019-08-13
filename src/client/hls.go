package client

import (
	"av"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"protocol/hls"
	"strconv"
	"sync"
	"time"
)

type HlsCache struct {
	Data      []byte
	SeqNumber int
	Duration  float64
}

func NewHlsCache(seq int, duration float64, data []byte) *HlsCache {
	cache := &HlsCache{
		Data:      data,
		SeqNumber: seq,
		Duration:  duration,
	}
	return cache
}

type Hls struct {
	isOld     bool
	uri       string
	token     string
	ts        *hls.TS
	seqNumber int
	startTime time.Time
	tsBuf     bytes.Buffer
	cache     sync.Map
}

func NewHlsClient(uri string) *Hls {
	hls := &Hls{
		startTime: time.Now(),
		uri:       uri,
		ts:        &hls.TS{},
	}
	hls.Token()
	return hls
}

func (H Hls) IsHls() (ret bool) {
	ret = true
	return
}

func (H *Hls) IsOld() (ret bool) {
	ret = H.isOld
	H.isOld = true
	return
}

func (H Hls) Uri() (ret string) {
	ret = H.uri
	return
}

func (H *Hls) Token() (ret string) {
	if H.token != "" {
		ret = H.token
		return
	}
	text := H.uri + "Hls" + time.Now().Format("2006-01-02 15:04:05")
	hasher := md5.New()
	hasher.Write([]byte(text))
	H.token = hex.EncodeToString(hasher.Sum(nil))
	ret = H.token
	return
}

func (H *Hls) ReceivePacket(packet *av.Packet) {
	switch {
	case packet.IsAudioSequence(), packet.IsVideoSequence():
		H.ts.ReadSeq(packet)
	case packet.IsPureAudioData(), packet.IsPureVideoData():
		if packet.IsVideoKeyFrame() {
			if H.tsBuf.Len() != 0 {
				duration := time.Now().Sub(H.startTime).Seconds()
				cache := NewHlsCache(H.seqNumber, duration, H.tsBuf.Bytes())
				index := strconv.Itoa(H.seqNumber) + ".ts"
				H.cache.Store(index, cache)
				H.tsBuf = bytes.Buffer{}
				H.seqNumber++
			}
			H.tsBuf.Write(hls.GetPAT())
			H.tsBuf.Write(H.ts.GetPMT())
			H.startTime = time.Now()
		}
		H.tsBuf.Write(H.ts.ReadData(packet))
	}
}

func (H *Hls) Close() {
}

func (H Hls) GetPacket() (packet *av.Packet, ok bool) {
	return
}

func (H *Hls) WriteM3u8(name string, w http.ResponseWriter) {
	body := bytes.NewBuffer(nil)
	getSeq, length := false, 0
	H.cache.Range(func(_, _ interface{}) bool {
		length++
		return true
	})
	for index := 0; index < length; index++ {
		ts := strconv.Itoa(index) + ".ts"
		if val, ok := H.cache.Load(ts); ok && val != nil {
			cache := val.(*HlsCache)
			if !getSeq {
				getSeq = true
				fmt.Fprintf(body,
					"#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-ALLOW-CACHE:NO\n#EXT-X-TARGETDURATION:%d\n#EXT-X-MEDIA-SEQUENCE:%d\n\n",
					4, cache.SeqNumber)
			}
			fmt.Fprintf(body, "#EXTINF:%.3f\n%s\n", cache.Duration, "/"+name+"/"+ts)

		}
	}
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "application/x-mpegURL")
	w.Header().Set("Content-Length", strconv.Itoa(body.Len()))
	w.Write(body.Bytes())

}

func (H *Hls) WriteTs(index string, w http.ResponseWriter) {
	if val, ok := H.cache.Load(index); ok {
		cache := val.(*HlsCache)
		w.Header().Set("Content-Type", "video/mp2ts")
		w.Header().Set("Content-Length", strconv.Itoa(len(cache.Data)))
		w.Write(cache.Data)
	}
}
