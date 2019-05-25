package flv

import (
	"bits"
	"math"
	"reflect"
	"time"
)

func TsToTime(ts uint32) time.Duration {
	return time.Millisecond * time.Duration(ts)
}

func TimeToTs(tm time.Duration) uint32 {
	return uint32(tm / time.Millisecond)
}

type AMFMap map[string]interface{}

const (
	amfNumber    = iota //double类型
	amfBoolean          //bool类型
	amfString           //string类型
	amfObject           //object类型
	amfMovieClip        //	Not available in Remoting
	amfNull             //null类型，空
	amfUndefined
	amfReference
	amfMixedArray
	amfEndOfObject //See Object ，表示object结束
	amfArray
	amfDate
	amfLongString
	amfUnsupported
	amfRecordset //Remoting, server-to-client only
	amfXML
	amfTypedObject //(Class instance)
	amfAMF3Data    //Sent by Flash player 9+
)

func ParserDataMsgAMF0(data []byte) []interface{} {
	val := make([]interface{}, 0)
	for {
		if len(data) == 0 {
			break
		}
		offset, _val := ParserAMFData(data)
		val = append(val, _val)
		data = data[offset:]
	}
	return val

}

func ParserCommandMsgAMF0(data []byte) []interface{} {
	return ParserDataMsgAMF0(data)
}

func ParserAMFData(input []byte) (offset int, val interface{}) {
	if len(input) == 0 {
		return 0, nil
	}
	amfType := input[0]
	offset++
	switch int(amfType) {
	case amfNumber:
		_offset, _val := parserAMFNumber(input[1:9])
		offset += _offset
		val = _val
	case amfBoolean:
		_offset, _val := parserAMFBoolean(input[1])
		offset += _offset
		val = _val
	case amfString:
		_offset, _val := parserAMFString(input[1:])
		offset += _offset
		val = _val
	case amfObject:
		data := input[1:]
		_offset, _val := parserAMFObject(data)
		offset += _offset
		val = _val
	case amfNull:
		val = nil
	case amfMixedArray:
		data := input[1:]
		offset += 4
		data = data[4:]
		_offset, _val := parserAMFObject(data)
		offset += _offset
		val = _val
	}
	return
}

func parserAMFString(data []byte) (offset int, val string) {
	length := bits.U16BE(data[:2])

	offset = 2 + int(length)
	val = string(data[2:offset])
	return
}

func parserAMFNumber(data []byte) (offset int, val float64) {
	// offset 要加上 type
	return 8, math.Float64frombits(bits.U64BE(data[:8]))
}

func parserAMFBoolean(data byte) (offset int, val bool) {
	offset = 1
	if data == 0x00 {
		val = false
		return
	}
	val = true
	return
}
func parserAMFObject(data []byte) (offset int, amfObject AMFMap) {
	amfObject = AMFMap{}
	for {
		if bits.U24BE(data[:3]) == 9 {
			offset += 3
			break
		}
		_offset, _val := ParserAMFData(append([]byte{0x02}, data...))
		offset += _offset - 1
		key := _val.(string)
		data = data[_offset-1:]
		_offset, _val = ParserAMFData(data)
		offset += _offset
		amfObject[key] = _val
		data = data[_offset:]
	}
	return
}

func AMFValLen(_val interface{}) (length int) {
	length = 1
	switch val := _val.(type) {
	case int8, int16, int32, int64, int, uint8, uint16,
		uint32, uint64, uint, float32, float64:
		length += 8
	case string:
		length += len(val)
		length += 2
	case AMFMap:
		for k, v := range val {
			length += len(k) + 2
			length += AMFValLen(v)
		}
		length += 3

	case bool:
		length += 1
	case nil:
		length += 0
	}
	return
}

func AMFValFillNumber(_val interface{}) (data []byte) {
	data = make([]byte, 9)
	data[0] = 0x00
	var input float64
	val := reflect.ValueOf(_val)
	switch _val.(type) {
	case int, int8, int16, int32, int64:
		input = float64(val.Int())

	case uint, uint8, uint16, uint32, uint64:
		input = float64(val.Uint())

	case float32, float64:
		input = float64(val.Float())

	}
	copy(data[1:], bits.PutU64BE(math.Float64bits(input)))
	return
}

func AMFValFill(_val interface{}) (data []byte) {
	switch val := _val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		data = AMFValFillNumber(_val)
	case string:
		length := len(val)
		data = make([]byte, length+3)
		data[0] = 0x02
		copy(data[1:3], bits.PutU16BE(uint16(length)))
		copy(data[3:], []byte(val))
	case AMFMap:
		data = []byte{0x03}
		for k, v := range val {
			length := bits.PutU16BE(uint16(len(k)))
			data = append(data, length...)
			data = append(data, []byte(k)...)
			data = append(data, AMFValFill(v)...)
		}
		data = append(data, []byte{0x00, 0x00, 0x09}...)
	case bool:
		data = make([]byte, 2)
		data[0] = 0x01
		if val {
			data[1] = 0x01
		} else {
			data[1] = 0x00
		}

	case nil:
		data = []byte{0x05}
	}
	return
}
