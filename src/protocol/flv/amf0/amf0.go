package amf0

import (
	"bytes"
)

type Amf0Map map[string]interface{}

var (
	Amf0EndFlag = []byte{0x00, 0x00, 0x09}
	AMf0Publish = Amf0Map{
		"level":       "status",
		"code":        "NetStream.Publish.Start",
		"description": "Start publishing",
	}
	AMf0Play = Amf0Map{
		"level":       "status",
		"code":        "NetStream.Play.Start",
		"description": "Start playing stream.",
	}
	Amf0ConnectOpt0 = Amf0Map{
		"fmtVer":       "FMS/3,0,1,123",
		"capabilities": 31,
	}
	Amf0ConnectOpt1 = Amf0Map{
		"level":          "status",
		"code":           "NetConnection.Connect.Success",
		"description":    "Connection succeeded.",
		"objectEncoding": 0,
	}
	AMf0Seek = Amf0Map{
		"level":       "status",
		"code":        "NetStream.Seek.Notify",
		"description": "Seek succeeded.",
	}
)

const (
	amf0Number    = iota //double类型
	amf0Boolean          //bool类型
	amf0String           //string类型
	amf0Object           //object类型
	amf0MovieClip        //	Not available in Remoting
	amf0Null             //null类型，空
	amf0Undefined
	amf0Reference
	amf0MixedArray
	amf0EndOfObject //See Object ，表示object结束
	amf0Array
	amf0Date
	amf0LongString
	amf0Unsupported
	amf0Recordset //Remoting, server-to-client only
	amf0XML
	amf0TypedObject //(Class instance)
	amf0AMF3Data    //Sent by Flash player 9+
)

type Amf0Cmd struct {
	CmdName string
	TransID float64
	Obj     Amf0Map
	Opts    []interface{}
}

func NewAmf0FromBytes(data []byte) (ret *Amf0Cmd) {
	ret = &Amf0Cmd{}
	ret.Decode(data)
	return
}

func NewAmf0(cmdName string, transid float64, obj Amf0Map, opts ...interface{}) (ret Amf0Cmd) {
	ret = Amf0Cmd{
		CmdName: cmdName,
		TransID: transid,
		Obj:     obj,
		Opts:    opts,
	}
	return
}

func (A *Amf0Cmd) Decode(data []byte) {
	data, A.CmdName = Amf0DecodeString(data[1:])
	data, A.TransID = Amf0DecodeNumber(data[1:])
	if data[0] == amf0Object {
		data, A.Obj = Amf0DecodeObject(data[1:])
	} else {
		data = data[1:]
	}
	A.Opts = Amf0DecodeMsg(data)
}

func (A *Amf0Cmd) Encode() (data []byte) {
	var tmp bytes.Buffer
	tmp.Write(Amf0EncodeString(A.CmdName))
	tmp.Write(Amf0EncodeNumber(A.TransID))
	tmp.Write(Amf0EncodeObject(A.Obj))
	for _, val := range A.Opts {
		tmp.Write(Amf0Encode(val))
	}
	data = tmp.Bytes()
	return
}
