package client

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
)

var (
	FMSKey = []byte{
		0x47, 0x65, 0x6e, 0x75, 0x69, 0x6e, 0x65, 0x20,
		0x41, 0x64, 0x6f, 0x62, 0x65, 0x20, 0x46, 0x6c,
		0x61, 0x73, 0x68, 0x20, 0x4d, 0x65, 0x64, 0x69,
		0x61, 0x20, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72,
		0x20, 0x30, 0x30, 0x31, // Genuine Adobe Flash Media Server 001
		0xf0, 0xee, 0xc2, 0x4a, 0x80, 0x68, 0xbe, 0xe8,
		0x2e, 0x00, 0xd0, 0xd1, 0x02, 0x9e, 0x7e, 0x57,
		0x6e, 0xec, 0x5d, 0x2d, 0x29, 0x80, 0x6f, 0xab,
		0x93, 0xb8, 0xe6, 0x36, 0xcf, 0xeb, 0x31, 0xae,
	}
	FPKey = []byte{
		0x47, 0x65, 0x6E, 0x75, 0x69, 0x6E, 0x65, 0x20,
		0x41, 0x64, 0x6F, 0x62, 0x65, 0x20, 0x46, 0x6C,
		0x61, 0x73, 0x68, 0x20, 0x50, 0x6C, 0x61, 0x79,
		0x65, 0x72, 0x20, 0x30, 0x30, 0x31, // Genuine Adobe Flash Player 001
		0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8,
		0x2E, 0x00, 0xD0, 0xD1, 0x02, 0x9E, 0x7E, 0x57,
		0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
		0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	}
	FMSVersion = []byte{0x04, 0x05, 0x00, 0x01}
)

const (
	RtmpServerVersion = 3
	DigestOffsetLen   = 4
	DigestLen         = 32
	DigestMaxOffset   = 728 //764 - 32 - 4
	Schema0Base       = 772
	Schema1Base       = 8
)

func prepareHandshakeData() (C1, C2, C0C1, S1, S2, S0S1S2 []byte) {
	var random [(1 + 1536*2) * 2]byte
	C0C1C2 := random[:1+1536*2]
	C0C1 = C0C1C2[:1536+1]
	C1 = C0C1C2[1 : 1536+1]
	C2 = C0C1C2[1536+1:]
	S0S1S2 = random[1+1536*2:]
	S0 := S0S1S2[:1]
	S1 = S0S1S2[1 : 1536+1]
	S2 = S0S1S2[1536+1:]
	S0[0] = RtmpServerVersion
	return
}

func simpleHandshake(C1, S1, S2 []byte) (ok bool) {
	clientVersion := binary.BigEndian.Uint32(C1[4:8])
	if clientVersion == 0 {
		copy(S1, C1)
		copy(S2, C1)
		ok = true
	}
	return
}

func complexHandshake(C1, S1, S2 []byte) (digest []byte, ok bool) {
	digest, ok = getDigest(C1)
	clientTime := C1[:4]
	makeS1(S1, clientTime)
	makeS2(S2, digest)
	return
}

func getDigest(C1 []byte) (digest []byte, ok bool) {
	if digest, ok = findSchema(C1, FPKey[:30], Schema1Base); !ok {
		digest, ok = findSchema(C1, FPKey[:30], Schema0Base)
	}

	return
}

/*
					c1s1 schema0:
									time: 4bytes
									version: 4bytes
									key: 764bytes
									digest: 764bytes

					c1s1 schema1:
									time: 4bytes
									version: 4bytes
									digest: 764bytes
									key: 764bytes
************************************************************************************************
					764bytes key struct:
								random-data: (offset)bytes
								key-data: 128bytes
								random-data: (764-offset-128-4)bytes
								offset: 4bytes

					764bytes digest struct:
								offset: 4bytes
								random-data: (offset)bytes
								digest-data: 32bytes
								random-data: (764-4-offset-32)bytes
*/
func findSchema(data, key []byte, base int) (digest []byte, ok bool) {
	offset := getDigestOffset(data, base)
	digest = calcDigest(data[:offset], data[offset+DigestLen:], key)
	ok = bytes.Compare(digest, data[offset:offset+DigestLen]) == 0
	return
}

func getDigestOffset(data []byte, base int) (offset int) {
	for _, val := range data[base : base+DigestOffsetLen] {
		offset += int(val)
	}
	offset = base + DigestOffsetLen + (offset % DigestMaxOffset)
	return
}

func calcDigest(part1, part2, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(part1)
	h.Write(part2)
	return h.Sum(nil)
}

func makeS1(S1, timestamp []byte) {
	rand.Read(S1)
	copy(S1[:4], timestamp)
	copy(S1[4:8], FMSVersion)
	digest, _ := findSchema(S1, FMSKey[:36], Schema1Base)
	offset := getDigestOffset(S1, Schema1Base)
	copy(S1[offset:offset+DigestLen], digest)
}

/*
	1536bytes C2S2 struct:
						random-data: 1504bytes
						digest-data: 32bytes
*/
func makeS2(S2, c1Digest []byte) {
	rand.Read(S2)
	randomDataLen := len(S2) - DigestLen
	tempKey := calcDigest(c1Digest, nil, FMSKey)
	s2Digest := calcDigest(S2[:randomDataLen], nil, tempKey)
	copy(S2[randomDataLen:], s2Digest)
}
