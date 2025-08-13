package ws

import (
	"encoding/binary"
	"io"
)

type frameHeader struct {
	Fin        bool
	Type       wsFrameType
	Rsv1Flag   bool
	Rsv2Flag   bool
	Rsv3Flag   bool
	Length     int
	MaskingKey []byte
}

func readFrameHeader(reader io.Reader) (*frameHeader, error) {
	var firstTwo [2]byte
	_, err := io.ReadFull(reader, firstTwo[:])
	if err != nil {
		return nil, err
	}

	first := firstTwo[0]
	maskAndLen := firstTwo[1]

	var header frameHeader
	header.Fin = (first & 0x80) != 0
	header.Type = wsFrameType(first & 0x0F)
	header.Rsv1Flag = (first & 0x40) != 0
	header.Rsv2Flag = (first & 0x20) != 0
	header.Rsv3Flag = (first & 0x10) != 0

	masked := (maskAndLen & 0x80) != 0
	header.Length = int(maskAndLen & 0x7F)

	switch header.Length {
	case 126:
		var ext [2]byte
		if _, err = io.ReadFull(reader, ext[:]); err != nil {
			return nil, err
		}
		header.Length = int(binary.BigEndian.Uint16(ext[:]))
	case 127:
		var ext [8]byte
		if _, err = io.ReadFull(reader, ext[:]); err != nil {
			return nil, err
		}
		header.Length = int(binary.BigEndian.Uint64(ext[:]))
	}

	if masked {
		header.MaskingKey = make([]byte, 4)
		if _, err = io.ReadFull(reader, header.MaskingKey); err != nil {
			return nil, err
		}
	}

	return &header, nil
}

func prepareFrameHeader(header *frameHeader) []byte {
	first := byte(header.Type & 0x0F)
	if header.Fin {
		first |= 0x80
	}
	if header.Rsv1Flag {
		first |= 0x40
	}
	if header.Rsv2Flag {
		first |= 0x20
	}
	if header.Rsv3Flag {
		first |= 0x10
	}

	buf := []byte{first}

	var maskBit byte
	if header.MaskingKey != nil {
		maskBit = 0x80
	}

	switch {
	case header.Length <= 125:
		buf = append(buf, byte(header.Length)|maskBit)
	case header.Length < 65536:
		buf = append(buf, 126|maskBit)
		var ext [2]byte
		binary.BigEndian.PutUint16(ext[:], uint16(header.Length))
		buf = append(buf, ext[:]...)
	default:
		buf = append(buf, 127|maskBit)
		var ext [8]byte
		binary.BigEndian.PutUint64(ext[:], uint64(header.Length))
		buf = append(buf, ext[:]...)
	}

	if header.MaskingKey != nil {
		buf = append(buf, header.MaskingKey...)
	}
	return buf
}
