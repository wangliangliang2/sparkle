package mp4

import (
	"encoding/binary"
	"errors"
)

const (
	BoxHeaderSize   = 8
	SoundFormat_AAC = 10
)

const (
	SoundSampleRate_5_5Khz = iota
	SoundSampleRate_11Khz
	SoundSampleRate_22Khz
	SoundSampleRate_44Khz
)

const (
	SoundSampleSize_8BIT = iota
	SoundSampleSize_16BIT
)

const (
	SoundSample_MONO = iota
	SoundSample_STEREO
)

const (
	MovTimescale                 = 1000
	MovTimescaleToMediaTimescale = 30
)

func Mp4Uint32BE(v uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return b
}

func Mp4Uint64BE(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

type AACInfo struct {
	SampleSize    uint32
	SampleRate    uint32
	SampleChannel uint32
}

func AudioAACInfo(audio []byte) (info AACInfo, err error) {
	flags := audio[0]

	SoundFormat := flags >> 4 & 0xf
	if SoundFormat == SoundFormat_AAC {
		switch (flags >> 2) & 0x3 {
		case SoundSampleRate_5_5Khz:
			info.SampleRate = 5500
		case SoundSampleRate_11Khz:
			info.SampleRate = 11025
		case SoundSampleRate_22Khz:
			info.SampleRate = 22050
		case SoundSampleRate_44Khz:
			info.SampleRate = 44100

		}
		switch (flags >> 1) & 0x1 {
		case SoundSampleSize_8BIT:
			info.SampleSize = 8
		case SoundSampleSize_16BIT:
			info.SampleSize = 16
		}

		switch flags & 0x1 {
		case SoundSample_MONO:
			info.SampleChannel = 1
		case SoundSample_STEREO:
			info.SampleChannel = 2
		}
		return
	}
	err = errors.New("not AAC")
	return
}

func GetFtypBox() Ftyp {
	return Ftyp{
		MajorBrand:       "isom",
		MinorVersion:     512,
		CompatibleBrands: "isomiso2avc1mp41",
	}
}

func GetFreeBox() Free {
	return Free{}
}

func GetMvhdBox(duration uint32) Mvhd {
	return Mvhd{
		TimeScale: MovTimescale,
		Duration:  duration * MovTimescale,
	}
}

func GetTkhdBox(isVideo bool, duration, width, height uint32) Tkhd {
	return Tkhd{
		Duration: duration * MovTimescale,
		IsVideo:  isVideo,
		Width:    width,
		Height:   height,
	}
}

func GetStcoBox(stco []byte) Stco {
	return Stco{
		NumberOfEntries:  uint32(len(stco) / 4),
		ChunkOffsetTable: stco,
	}
}

func GetStszBox(stsz []byte) Stsz {
	return Stsz{
		SampleSize:      0,
		NumberOfEntries: uint32(len(stsz) / 4),
		SampleSizeTable: stsz,
	}
}

func GetStscBox(stsc []byte) Stsc {
	return Stsc{
		NumberOfEntries:    uint32(len(stsc) / 12),
		SampleToChunkTable: stsc,
	}
}

func GetCttsBox(ctts []byte) Ctts {
	return Ctts{
		EntryCount:             uint32(len(ctts) / 8),
		CompositionOffsetTable: ctts,
	}
}

func GetStssBox(stss []byte) Stss {
	return Stss{
		NumberOfEntries: uint32(len(stss) / 4),
		SyncSampleTable: stss,
	}
}

func GetSttsBox(stts []byte) Stts {
	return Stts{
		NumberOfEntries:   1,
		TimeToSampleTable: stts,
	}
}

func GetVideoStsdBox(width, height uint32, seq []byte) Stsd {
	return Stsd{
		Avc1: Avc1{
			Width:  width,
			Height: height,
			AvcC: AvcC{
				VideoSeq: seq,
			},
		},
		IsVideo: true,
	}
}

func GetAudioStsdBox(aacInfo AACInfo, seq []byte) Stsd {

	return Stsd{
		IsVideo: false,
		Mp4a: Mp4a{
			Channels:   aacInfo.SampleChannel,
			SampleSize: aacInfo.SampleSize,
			SampleRate: aacInfo.SampleRate,
			Esds: Esds{
				Audioseq: seq,
			},
		},
	}
}

func GetVideoStblBox(width, height uint32, seq, stts, stss, ctts, stsc, stsz, stco []byte) Stbl {
	return Stbl{
		IsVideo: true,
		Stsd:    GetVideoStsdBox(width, height, seq),
		Stts:    GetSttsBox(stts),
		Stss:    GetStssBox(stss),
		Ctts:    GetCttsBox(ctts),
		Stsc:    GetStscBox(stsc),
		Stsz:    GetStszBox(stsz),
		Stco:    GetStcoBox(stco),
	}
}

func GetVideoTrak(duration, width, height uint32, seq, stts, stss, ctts, stsc, stsz, stco []byte) Trak {
	return Trak{
		Tkhd: GetTkhdBox(true, duration, width, height),
		Mdia: Mdia{
			Mdhd: Mdhd{
				TimeScale: MovTimescale * MovTimescaleToMediaTimescale,
				Duration:  duration * MovTimescale * MovTimescaleToMediaTimescale,
			},
			Hdlr: Hdlr{
				IsVideo: true,
				Name:    "VideoHandler",
			},
			Minf: Minf{
				IsVideo: true,
				Vmhd:    DefaultVmhd(),
				Dinf: Dinf{
					Dref: DefaultDref(),
				},
				Stbl: GetVideoStblBox(width, height, seq, stts, stss, ctts, stsc, stsz, stco),
			},
		},
	}

}

func GetAudioTrak(aacInfo AACInfo, duration uint32, seq, stts, stsc, stsz, stco []byte) Trak {
	return Trak{
		Tkhd: GetTkhdBox(false, duration, 0, 0),
		Mdia: Mdia{
			Mdhd: Mdhd{
				TimeScale: aacInfo.SampleRate,
				Duration:  duration * aacInfo.SampleRate,
			},
			Hdlr: Hdlr{
				IsVideo: false,
				Name:    "SoundHandler",
			},
			Minf: Minf{
				IsVideo: false,
				Smhd:    DefaultSmhd(),
				Dinf: Dinf{
					Dref: DefaultDref(),
				},
				Stbl: GetAudioStblBox(aacInfo, seq, stts, stsc, stsz, stco),
			},
		},
	}
}

func GetAudioStblBox(aacInfo AACInfo, seq, stts, stsc, stsz, stco []byte) Stbl {
	return Stbl{
		IsVideo: false,
		Stsd:    GetAudioStsdBox(aacInfo, seq),
		Stts:    GetSttsBox(stts),
		Stsc:    GetStscBox(stsc),
		Stsz:    GetStszBox(stsz),
		Stco:    GetStcoBox(stco),
	}
}
