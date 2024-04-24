package datatype

import (
	"encoding/binary"
	"math"
)

type encode struct{}

func (_ encode) Integer(value int32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf[:], uint32(value))
	return buf
}

func (_ encode) String(value string) []byte {
	// TODO: bounds checking on length of string
	buf := make([]byte, len(value)+2)
	binary.BigEndian.PutUint16(buf[:], uint16(len(value)))
	copy(buf[2:], []byte(value))
	return buf
}

func (_ encode) BigInteger(value int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf[:], uint64(value))
	return buf
}

func (_ encode) Boolean(value bool) []byte {
	buf := make([]byte, 1)
	if value {
		buf[0] = 1
	} else {
		buf[0] = 0
	}

	return buf
}

func (_ encode) Float(value float32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf[:], math.Float32bits(value))
	return buf
}