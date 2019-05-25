package server

import (
	"client"
	"log"
	"net/http"
	"path"
	"strings"
)

type HlsServer struct {
	Handle  Server
	clients map[string]*client.HlsClient
}

func NewHlsServer(handle Server) *HlsServer {
	if handle == nil {
		log.Fatalln("Need RTMP Server")
	}
	return &HlsServer{
		Handle:  handle,
		clients: make(map[string]*client.HlsClient),
	}
}

func (H *HlsServer) Serve() {
	go H.ManageClient()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		H.handleConnect(w, r)
	})
	http.ListenAndServe(":3333", mux)
}

func (H *HlsServer) ManageClient() {
	for {
		programs := H.Handle.GetPrograms()
		for _, programName := range programs {
			if _, ok := H.clients[programName]; !ok {
				subscriber := client.NewHlsClient()
				H.clients[programName] = subscriber
				H.Handle.AddClient(programName, subscriber)
			}
		}
		for key, client := range H.clients {
			if client.IsClosed() {
				delete(H.clients, key)
			}
		}
	}

}

func (H *HlsServer) handleConnect(w http.ResponseWriter, r *http.Request) {
	switch path.Ext(r.URL.Path) {
	case ".ts":
		components := strings.Split(r.URL.Path[1:], "/")
		programName := strings.Join(components[:len(components)-1], "/")
		ts := components[len(components)-1]
		if subscriber, ok := H.clients[programName]; ok {
			subscriber.WriteTs(ts, w)
		}
	case ".m3u8":
		programName := r.URL.Path[1 : len(r.URL.Path)-5]
		if subscriber, ok := H.clients[programName]; ok {
			subscriber.WriteM3u8(programName, w)
		}
	}
}
