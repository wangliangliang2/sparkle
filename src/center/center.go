package center

import (
	"av"
	"client"
	"sync"
)

type Show struct {
	Pub      client.Client
	Subs     sync.Map
	Cache    *Cache
	HlsToken string
}

type Center struct {
	list sync.Map
}

var defaultCenter = Center{}

func StartShow(cli client.Client) (ok bool) {
	ok = defaultCenter.startShow(cli)
	return
}

func JoinShow(cli client.Client) (ok bool) {
	ok = defaultCenter.joinShow(cli)
	return

}

func LeaveShow(cli client.Client) {
	defaultCenter.leaveShow(cli)
}

func GetHlsClient(uri string) (ret client.Client) {
	ret = defaultCenter.getHlsClient(uri)
	return
}

func (C *Center) startShow(cli client.Client) (ok bool) {
	if _, exist := C.list.Load(cli.Uri()); !exist {
		ok = true
		show := &Show{
			Pub:   cli,
			Cache: NewCache(),
		}
		C.list.Store(cli.Uri(), show)

		go show.start()
	}

	return
}

func (C *Center) joinShow(cli client.Client) (ok bool) {
	if val, exist := C.list.Load(cli.Uri()); exist {
		ok = true
		show := val.(*Show)
		go show.joinShow(cli)
	}
	return
}

func (C *Center) leaveShow(cli client.Client) {
	if val, exist := C.list.Load(cli.Uri()); exist {
		show := val.(*Show)
		if show.Pub.Token() == cli.Token() {
			show.Subs.Range(func(k, v interface{}) (ret bool) {
				ret = true
				sub := v.(client.Client)
				sub.Close()
				return
			})
			C.list.Delete(cli.Uri())
			return
		}
		go show.leave(cli)
	}
}

func (C *Center) getHlsClient(uri string) (cli client.Client) {
	if val, exist := C.list.Load(uri); exist {
		show := val.(*Show)
		if val, ok := show.Subs.Load(show.HlsToken); ok {
			cli = val.(client.Client)
		}
	}
	return
}

func (S *Show) start() {
	for {
		packet, ok := S.Pub.GetPacket()
		if !ok {
			LeaveShow(S.Pub)
			break
		}
		S.Cache.Save(packet)
		S.sendPacket(packet)
	}
}

func (S Show) sendPacket(packet *av.Packet) {
	S.Subs.Range(func(k, val interface{}) (ret bool) {
		ret = true
		cli := val.(client.Client)
		if !cli.IsOld() {
			S.Cache.Send(cli)
			return
		}
		cli.ReceivePacket(packet)
		return
	})
}

func (S *Show) leave(cli client.Client) {
	if val, ok := S.Subs.Load(cli.Token()); ok {
		cli := val.(client.Client)
		cli.Close()
		S.Subs.Delete(cli.Token())
	}
}

func (S *Show) joinShow(cli client.Client) {
	if cli.IsHls() {
		S.HlsToken = cli.Token()
	}
	S.Subs.Store(cli.Token(), cli)
}
