package server

import (
	"center"
	"client"
	"log"
	"net/http"
	"path"
	"strings"
)

const HlsListenPort = ":3333"

type Hls struct {
}

func NewHlsServer() Hls {
	return Hls{}
}

func (H Hls) Serve() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		H.handleConnect(w, r)
	})
	log.Println("Hls Server Listening ", HlsListenPort)
	http.ListenAndServe(HlsListenPort, mux)
}

func (H Hls) handleConnect(w http.ResponseWriter, r *http.Request) {
	switch path.Ext(r.URL.Path) {
	case ".ts":
		components := strings.Split(r.URL.Path[1:], "/")
		programName := strings.Join(components[:len(components)-1], "/")
		ts := components[len(components)-1]
		if cli := center.GetHlsClient(programName); cli != nil {
			hls := cli.(*client.Hls)
			hls.WriteTs(ts, w)
		}
	case ".m3u8":
		programName := r.URL.Path[1 : len(r.URL.Path)-5]
		if cli := center.GetHlsClient(programName); cli != nil {
			hls := cli.(*client.Hls)
			hls.WriteM3u8(programName, w)
		}

	}
}
