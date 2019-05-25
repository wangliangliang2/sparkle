package main

import (
	"log"
	"server"
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)
	rtmpServer := server.NewRtmpServer()
	flvServer := server.NewFlvServer(rtmpServer)
	go flvServer.Serve()
	hlsServer := server.NewHlsServer(rtmpServer)
	go hlsServer.Serve()
	rtmpServer.Serve()
}
