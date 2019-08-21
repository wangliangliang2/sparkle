package av

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
)

func naluCompetitionPrevention(in []byte) (ret []byte) {
	tmp := make([]byte, len(in))
	copy(tmp, in)
	for index := 0; index < len(tmp)-2; index++ {
		val := tmp[index] ^ 0x00 + tmp[index+1] ^ 0x00 + tmp[index+2] ^ 0x03
		if val == 0 {
			tmp = append(tmp[:index+2], tmp[index+3:]...)
		}
	}
	ret = tmp
	return
}

/*

	seq should be like the following:
	0x1, 0x64, 0x0, 0x1f, 0xff, 0xe1, 0x0, 0x19,
	0x67, 0x64, 0x0, 0x1f, 0xac, 0xc8, 0x60, 0x2d,
	0x2, 0x86, 0x84, 0x0, 0x0, 0x3, 0x0, 0x4, 0x0,
	0x0, 0x3, 0x0, 0xc8, 0x3c, 0x60, 0xc6, 0x68, 0x1,
	0x0, 0x5, 0x68, 0xe9, 0x78, 0xbc, 0xb0
*/
func getSPSFromVideoSeq(seq []byte) (ret []byte) {
	spsSize := binary.BigEndian.Uint16(seq[6:8])
	sps := seq[8 : 8+spsSize]
	ret = naluCompetitionPrevention(sps)
	return
}
func getBitstreamFromSps(sps []byte) Bitstream {
	var tmp bytes.Buffer
	for _, item := range sps {
		tmp.WriteString(fmt.Sprintf("%08b", item))
	}
	return Bitstream(tmp.String())
}

func isSps(bitstream *Bitstream) bool {
	naluType := bitstream.ReadByte() & 0x1f

	if naluType != 0x07 {
		return false
	}
	return true
}

func GetSizeFromVideoSeq(seq []byte) (width, height uint32) {
	bitstream := getBitstreamFromSps(getSPSFromVideoSeq(seq))
	if !isSps(&bitstream) {
		return
	}
	profileIdc := bitstream.ReadByte()
	bitstream.DropNBit(16)
	bitstream.UnsignExpGolomb() //seq_parameter_set_id
	if profileIdc == 100 || profileIdc == 110 || profileIdc == 122 || profileIdc == 144 {
		chromaFormatIdc := bitstream.UnsignExpGolomb()
		if chromaFormatIdc == 3 {
			bitstream.DropNBit(1)
		}
		bitstream.UnsignExpGolomb() //bit_depth_luma_minus8
		bitstream.UnsignExpGolomb() // bit_depth_chroma_minus8
		bitstream.DropNBit(1)       //qpprime_y_zero_transform_bypass_flag
		seqScalingMatrixPresentFlag := bitstream.ReadBit()
		if seqScalingMatrixPresentFlag == 1 {
			bitstream.DropNBit(8)
		}
	}
	bitstream.UnsignExpGolomb() //log2_max_frame_num_minus4
	picOrderCntType := bitstream.UnsignExpGolomb()
	switch picOrderCntType {
	case 0:
		bitstream.UnsignExpGolomb() //log2_max_pic_order_cnt_lsb_minus4
	case 1:
		bitstream.DropNBit(1)     //delta_pic_order_always_zero_flag
		bitstream.SignExpGolomb() //offset_for_non_ref_pic
		bitstream.SignExpGolomb() //offset_for_top_to_bottom_field
		numRefFramesInPicOrderCntCycle := bitstream.UnsignExpGolomb()
		for index := 0; index < numRefFramesInPicOrderCntCycle; index++ {
			bitstream.SignExpGolomb()
		}
	}
	bitstream.UnsignExpGolomb() // num_ref_frames
	bitstream.DropNBit(1)       //gaps_in_frame_num_value_allowed_flag
	picWidthInMbsMinus1 := bitstream.UnsignExpGolomb()
	picHeightInMapUnitsMinus1 := bitstream.UnsignExpGolomb()

	frameMbsOnlyFlag := bitstream.ReadBit()
	if frameMbsOnlyFlag == 0 {
		bitstream.DropNBit(1)
	}
	bitstream.DropNBit(1) //direct_8x8_inference_flag
	frameCroppingFlag := bitstream.ReadBit()
	frameCropLeftOffset := 0
	frameCropRightOffset := 0
	frameCropTopOffset := 0
	frameCropBottomOffset := 0
	if frameCroppingFlag == 1 {
		frameCropLeftOffset = bitstream.UnsignExpGolomb()
		frameCropRightOffset = bitstream.UnsignExpGolomb()
		frameCropTopOffset = bitstream.UnsignExpGolomb()
		frameCropBottomOffset = bitstream.UnsignExpGolomb()
	}
	width = uint32((picWidthInMbsMinus1+1)*16 - frameCropLeftOffset*2 - frameCropRightOffset*2)
	height = uint32((2-frameMbsOnlyFlag)*(picHeightInMapUnitsMinus1+1)*16 - frameCropTopOffset*2 - frameCropBottomOffset*2)
	return
}

type Bitstream string

func (B *Bitstream) ReadBit() (ret int) {
	entity := (*B)
	tmp := entity[0]
	*B = entity[1:]
	if tmp == '1' {
		ret = 1
	}
	return
}

func (B *Bitstream) ReadNBit(n int) (ret int) {
	entity := (*B)
	*B = entity[n:]
	tmp, _ := strconv.ParseInt(string(entity[:n]), 2, 64)
	ret = int(tmp)
	return
}

func (B *Bitstream) ReadByte() (ret int) {
	return B.ReadNBit(8)
}

func (B *Bitstream) DropNBit(n int) {
	*B = (*B)[n:]
}

func (B *Bitstream) SignExpGolomb() (ret int) {
	zeroCounter := 0
	for {
		if B.Peek(1) == 1 {
			break
		}
		B.DropNBit(1)
		zeroCounter++
	}
	ret = B.ReadNBit(zeroCounter)
	flag := B.ReadBit()
	if flag == 1 {
		ret = -ret
	}
	return
}

func (B *Bitstream) UnsignExpGolomb() (ret int) {

	zeroCounter := 0
	for {
		if B.Peek(1) == 1 {
			break
		}
		B.DropNBit(1)
		zeroCounter++
	}
	ret = B.ReadNBit(zeroCounter+1) - 1
	return
}

func (B Bitstream) Peek(n int) (ret int) {
	tmp, _ := strconv.ParseInt(string(B[:n]), 2, 64)
	ret = int(tmp)
	return
}
