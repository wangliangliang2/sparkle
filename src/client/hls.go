package client

import (
	"av"
	"bytes"
	"fmt"
	"net/http"
	"protocol/hls"
	"strconv"
	"time"
)

type HlsClient struct {
	Data             chan av.Packet
	Cache            map[string]*HlsCache
	isFresh, isClose bool
}

func NewHlsClient() *HlsClient {
	hls := &HlsClient{
		Data:    make(chan av.Packet, 1024),
		Cache:   make(map[string]*HlsCache),
		isFresh: true,
	}
	go hls.Store()
	return hls
}

func (H *HlsClient) Pull() (p av.Packet, err error) {
	return
}

func (H *HlsClient) Push(p av.Packet) error {
	switch {
	case p.IsAudio, p.IsVideo:
		H.Data <- p
	}
	return nil
}

func (H *HlsClient) IsClosed() bool {
	return H.isClose
}

func (H *HlsClient) IsFresh() bool {
	defer func() { H.isFresh = false }()
	return H.isFresh
}

func (H *HlsClient) ShutDown() {
	close(H.Data)
}

func (H *HlsClient) Store() {
	var tsContent bytes.Buffer
	ts, SeqNumber, StartTime := hls.TS{}, 0, time.Now()
	for {
		p, ok := <-H.Data
		if !ok {
			break
		}
		switch {
		case p.IsAudioSequence(), p.IsVideoSequence():
			ts.ReadSeq(p)
		case p.IsPureAudioData(), p.IsPureVideoData():
			if p.IsVideoKeyFrame() {
				if tsContent.Len() != 0 {
					duration := time.Now().Sub(StartTime).Seconds()
					cache := NewHlsCache(SeqNumber, duration, tsContent.Bytes())
					index := strconv.Itoa(SeqNumber) + ".ts"
					H.Cache[index] = cache
					tsContent.Reset()
					SeqNumber++
				}
				tsContent.Write(hls.MakePAT())
				tsContent.Write(ts.MakePMT())
				StartTime = time.Now()
			}
			tsContent.Write(ts.ReadData(p))
		}
	}
	for key, _ := range H.Cache {
		delete(H.Cache, key)
	}
	H.Cache = nil
	H.isClose = true
}

func (H *HlsClient) WriteM3u8(name string, w http.ResponseWriter) {
	body := bytes.NewBuffer(nil)
	getSeq, length := false, len(H.Cache)
	for index := 0; index < length; index++ {
		ts := strconv.Itoa(index) + ".ts"
		cache := H.Cache[ts]
		if cache != nil {
			if !cache.IsTimeout(length) {
				if !getSeq {
					getSeq = true
					fmt.Fprintf(body,
						"#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-ALLOW-CACHE:NO\n#EXT-X-TARGETDURATION:%d\n#EXT-X-MEDIA-SEQUENCE:%d\n\n",
						4, cache.SeqNumber)
				}
				fmt.Fprintf(body, "#EXTINF:%.3f\n%s\n", cache.Duration, "/"+name+"/"+ts)
			} else {
				H.Cache[ts] = nil
			}
		}
	}
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "application/x-mpegURL")
	w.Header().Set("Content-Length", strconv.Itoa(body.Len()))
	w.Write(body.Bytes())

}
func (H *HlsClient) WriteTs(index string, w http.ResponseWriter) {
	if cache, ok := H.Cache[index]; ok {
		w.Header().Set("Content-Type", "video/mp2ts")
		w.Header().Set("Content-Length", strconv.Itoa(len(cache.Data)))
		w.Write(cache.Data)
	}
}

type HlsCache struct {
	Data      []byte
	SeqNumber int
	Time      time.Time
	Duration  float64
}

func (H *HlsCache) IsTimeout(lastIndex int) bool {
	return lastIndex-3-H.SeqNumber > 0
}

func NewHlsCache(seq int, duration float64, data []byte) *HlsCache {
	cache := &HlsCache{
		Data:      make([]byte, len(data)),
		SeqNumber: seq,
		Time:      time.Now(),
		Duration:  duration,
	}
	copy(cache.Data, data)
	return cache
}
