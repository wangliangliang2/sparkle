package main

import (
	"log"
	"server"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	go server.NewRtmpServer().Serve()
	go server.NewFlvServer().Serve()
	go server.NewHlsServer().Serve()
	select {}
}
