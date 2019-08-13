package server

import (
	"center"
	"client"
	"log"
	"net"
)

const RtmpListenPort = ":1935"

type Rtmp struct {
}

func NewRtmpServer() Rtmp {
	return Rtmp{}
}

func (R Rtmp) Serve() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	l, err := net.Listen("tcp", RtmpListenPort)
	if err != nil {
		panic("RTMP Server Can't Start...")
	}
	log.Println("RTMP Server Listening ", RtmpListenPort)
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		cli := client.NewRtmpClient(conn)
		R.handleConn(cli)
	}
}

func (R Rtmp) handleConn(cli *client.Rtmp) {
	go cli.Work()
	go R.addClientToCenter(cli)
}

func (R Rtmp) addClientToCenter(cli *client.Rtmp) {
	<-cli.Ready
	if cli.IsPub() {
		if ok := center.StartShow(cli); !ok {
			cli.CloseConn()
		}
		if ok := center.JoinShow(client.NewHlsClient(cli.Uri())); !ok {
			log.Println("hls client failed ....")
		}
		return
	}
	if ok := center.JoinShow(cli); !ok {
		cli.CloseConn()
	}
}
