package masp

import (
	"encoding/binary"
	"errors"
	"io"
)

func read(r io.Reader, n int) ([]byte, error) {
	b := make([]byte, n)
	l, err := r.Read(b)
	if l != n {
		return nil, errors.New("failed to read required bytes")
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}

func ReadCompactSize(r io.Reader) (int, error) {
	flags, err := read(r, 1)
	if err != nil {
		return 0, err
	}
	flag := flags[0]
	if flag < 253 {
		return int(flag), nil
	}
	switch flag {
	case 253:
		data, err := read(r, 2)
		if err != nil {
			return 0, err
		}
		return int(binary.LittleEndian.Uint16(data)), nil
	case 254:
		data, err := read(r, 3)
		if err != nil {
			return 0, err
		}
		return int(binary.LittleEndian.Uint32(data)), nil
	default:
		data, err := read(r, 4)
		if err != nil {
			return 0, err
		}
		return int(binary.LittleEndian.Uint64(data)), nil
	}
}
