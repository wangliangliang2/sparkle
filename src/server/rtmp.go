package server

import (
	"client"
	"log"
	"net"
	"program"
	"sync"
)

type RtmpServer struct {
	Programs map[string]*program.Program
	mu       sync.RWMutex
}

func (R *RtmpServer) GetPrograms() []string {
	programs := make([]string, 0, len(R.Programs))
	R.mu.RLock()
	defer R.mu.RUnlock()
	for programName, p := range R.Programs {
		if !p.IsClosed() {
			programs = append(programs, programName)
		}
	}
	return programs
}

func (R *RtmpServer) ExistProgram(name string) bool {
	if p, ok := R.Programs[name]; ok && !p.IsClosed() {
		return true
	}
	return false
}

func (R *RtmpServer) AddClient(name string, subscriber client.Client) {
	if p, ok := R.Programs[name]; ok && !p.IsClosed() {
		p.AddSubscriber(subscriber)
	}
}

func NewRtmpServer() *RtmpServer {
	return &RtmpServer{
		Programs: make(map[string]*program.Program),
	}
}

func (R *RtmpServer) Serve() {
	l, err := net.Listen("tcp", ":1935")
	if err != nil {
		log.Fatalln("can't use port 1935")
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		stream := client.NewRtmpClient(conn)
		go R.Handle(stream)
	}
}

func (R *RtmpServer) Handle(stream *client.RtmpClient) {
	if err := stream.Handshake(); err != nil {
		log.Println(err)
		return
	}
	programName, isPub := stream.Start()
	switch {
	case isPub && !R.ExistProgram(programName):
		p := program.NewProgram(programName, stream)
		R.mu.Lock()
		R.Programs[programName] = p
		R.mu.Unlock()
		stream.ReceiveData()
		p.Begin()
	case !isPub && R.ExistProgram(programName):
		R.AddClient(programName, stream)
		stream.ReceiveData()
	default:
		stream.ShutDown()
	}
}
