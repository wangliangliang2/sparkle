package amf0

import (
	"bytes"
	"encoding/binary"
	"math"
	"reflect"
)

func Amf0EncodeString(input string) (ret []byte) {
	data := []byte(input)
	length := make([]byte, 2)
	binary.BigEndian.PutUint16(length, uint16(len(data)))
	var tmp bytes.Buffer
	tmp.WriteByte(amf0String)
	tmp.Write(length)
	tmp.Write(data)
	ret = tmp.Bytes()
	return
}

func Amf0EncodeNumber(input float64) (ret []byte) {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, math.Float64bits(input))
	var tmp bytes.Buffer
	tmp.WriteByte(amf0Number)
	tmp.Write(data)
	ret = tmp.Bytes()
	return
}

func Amf0EncodeObject(input Amf0Map) (ret []byte) {
	if len(input) == 0 {
		ret = Amf0EncodeNil()
		return
	}
	var tmp bytes.Buffer
	tmp.WriteByte(amf0Object)
	for key, val := range input {
		data := []byte(key)
		length := make([]byte, 2)
		binary.BigEndian.PutUint16(length, uint16(len(data)))
		tmp.Write(length)
		tmp.Write(data)
		tmp.Write(Amf0Encode(val))
	}
	tmp.Write(Amf0EndFlag)
	ret = tmp.Bytes()
	return
}

func Amf0EncodeBoolean(input bool) (ret []byte) {
	if input {
		ret = []byte{amf0Boolean, 0x01}
		return
	}
	ret = []byte{amf0Boolean, 0x00}
	return
}

func Amf0EncodeNil() (ret []byte) {
	ret = []byte{amf0Null}
	return
}

func Amf0Encode(input interface{}) (ret []byte) {
	switch val := input.(type) {
	case int, int8, int16, int32, int64:
		_val := reflect.ValueOf(val)
		ret = Amf0EncodeNumber(float64(_val.Int()))
	case uint, uint8, uint16, uint32, uint64:
		_val := reflect.ValueOf(val)
		ret = Amf0EncodeNumber(float64(_val.Uint()))
	case float32, float64:
		_val := reflect.ValueOf(val)
		ret = Amf0EncodeNumber(_val.Float())
	case string:
		ret = Amf0EncodeString(val)
	case Amf0Map:
		ret = Amf0EncodeObject(val)
	case bool:
		ret = Amf0EncodeBoolean(val)
	case nil:
		ret = Amf0EncodeNil()
	}
	return
}
