package program

import (
	"av"
	"client"
)

type Program struct {
	Name        string
	Publisher   client.Client
	Subscribers map[client.Client]bool
	Cache       *Cache
	isClosed    bool
}

func NewProgram(name string, publisher client.Client) *Program {
	return &Program{
		Name:        name,
		Publisher:   publisher,
		Subscribers: make(map[client.Client]bool),
		Cache:       NewCache(),
	}
}

func (P *Program) Begin() {
	for {
		packet, err := P.GetPacketFromPublisher()
		if err != nil {
			P.End()
			break
		}
		P.CachePacket(packet)
		P.SendPacketToSubscriber(packet)
	}
}

func (P *Program) SendPacketToSubscriber(packet av.Packet) {
	for subscriber, _ := range P.Subscribers {
		if !subscriber.IsClosed() {
			if subscriber.IsFresh() {
				P.Cache.Send(subscriber)
			} else {
				subscriber.Push(packet)
			}
		} else {
			P.RemoveSubscriber(subscriber)
		}
	}
}

func (P *Program) GetPacketFromPublisher() (av.Packet, error) {
	return P.Publisher.Pull()
}

func (P *Program) CachePacket(packet av.Packet) {
	P.Cache.Save(packet)
}

func (P *Program) AddPublisher(publisher client.Client) {
	P.Publisher = publisher
}

func (P *Program) AddSubscriber(subscriber client.Client) {
	P.Subscribers[subscriber] = true
}

func (P *Program) RemoveSubscriber(subscriber client.Client) {
	subscriber.ShutDown()
	delete(P.Subscribers, subscriber)

}

func (P *Program) IsClosed() bool {
	return P.isClosed
}

func (P *Program) End() {
	P.Publisher = nil
	for subscriber, _ := range P.Subscribers {
		P.RemoveSubscriber(subscriber)
	}
	P.Subscribers = nil
	P.isClosed = true
}
