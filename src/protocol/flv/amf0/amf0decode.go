package amf0

import (
	"bytes"
	"encoding/binary"
	"math"
)

func Amf0DecodeMsg(input []byte) (ret []interface{}) {
	if len(input) == 0 {
		return
	}
	ret = make([]interface{}, 0)
	for {
		var val interface{}
		input, val = Amf0Decode(input)
		ret = append(ret, val)
		if len(input) == 0 {
			break
		}

	}
	return
}

func Amf0Decode(input []byte) (output []byte, ret interface{}) {

	amf0Type := input[0]
	input = input[1:]
	switch amf0Type {
	case amf0Number:
		output, ret = Amf0DecodeNumber(input)
	case amf0Boolean:
		output, ret = Amf0DecodeBoolean(input)
	case amf0String:
		output, ret = Amf0DecodeString(input)
	case amf0Null:
		output, ret = input, nil
	case amf0Object:
		output, ret = Amf0DecodeObject(input)
	case amf0MixedArray:
		output, ret = Amf0DecodeMixedArray(input)
	}
	return
}

func Amf0DecodeNumber(input []byte) (output []byte, ret float64) {
	ret = math.Float64frombits(binary.BigEndian.Uint64(input[0:8]))
	output = input[8:]
	return
}

func Amf0DecodeBoolean(input []byte) (output []byte, ret bool) {
	ret = input[0] > 0
	output = input[1:]
	return
}

func Amf0DecodeString(input []byte) (output []byte, ret string) {
	length := binary.BigEndian.Uint16(input[0:2])
	ret = string(input[2 : 2+length])
	output = input[2+length:]
	return
}

func Amf0DecodeObject(input []byte) (output []byte, ret Amf0Map) {
	ret = Amf0Map{}
	var key string
	var val interface{}
	for {
		input, key = Amf0DecodeString(input)
		input, val = Amf0Decode(input)
		ret[key] = val
		if bytes.Compare(input[:3], Amf0EndFlag) == 0 {
			output = input[3:]
			break
		}
	}
	return
}

func Amf0DecodeMixedArray(input []byte) (output []byte, ret Amf0Map) {
	input = input[3:]
	input[4] = amf0Object
	output, ret = Amf0DecodeObject(input)
	return
}
