package server

import (
	"center"
	"client"
	"log"
	"net/http"
)

const FlvListenPort = ":2222"

type Flv struct {
}

func NewFlvServer() Flv {
	return Flv{}
}

func (F Flv) Serve() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		F.handleConnect(w, r)
	})
	log.Println("FLV Server Listening ", FlvListenPort)
	http.ListenAndServe(FlvListenPort, mux)
}

func (F Flv) handleConnect(w http.ResponseWriter, r *http.Request) {
	flv := client.NewFlvClient("rtmplive/home", w)
	go F.joinShow(flv)
	flv.Work()

}

func (F Flv) joinShow(flv *client.Flv) {
	if ok := center.JoinShow(flv); !ok {
		flv.Close()
	}
	<-flv.Notice
	center.LeaveShow(flv)

}
