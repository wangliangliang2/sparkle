package server

import (
	"client"
	"log"
	"net/http"
)

type FlvServer struct {
	Handle Server
}

func NewFlvServer(handle Server) *FlvServer {
	if handle == nil {
		log.Fatalln("Need RTMP Server")
	}
	return &FlvServer{
		Handle: handle,
	}
}

func (F *FlvServer) Serve() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		F.handleConnect(w, r)
	})
	http.ListenAndServe(":2222", mux)

}

func (F *FlvServer) handleConnect(w http.ResponseWriter, r *http.Request) {
	programName := r.URL.Path[1:]
	if F.Handle.ExistProgram(programName) {
		subscriber := client.NewFlvClient(w)
		F.Handle.AddClient(programName, subscriber)
		subscriber.Play()
	}
}
