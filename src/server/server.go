package server

import (
	"client"
)

type Server interface {
	ExistProgram(name string) bool
	AddClient(program string, subscriber client.Client)
	GetPrograms() []string
}
