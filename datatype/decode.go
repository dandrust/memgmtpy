package datatype

import (
	"encoding/binary"
	"math"
)

type decode struct{}

func (_ decode) Integer(buf []byte) int32 {
	return int32(binary.BigEndian.Uint32(buf))
}

func (_ decode) String(buf []byte) string {
	end := binary.BigEndian.Uint16(buf[:])
	return string(buf[2:2 + end])
}

func (_ decode) BigInteger(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

func (_ decode) Boolean(buf []byte) bool {
	var out bool

	if buf[0] == byte(0) {
		out = false
	} else {
		out = true
	}

	return out
}

func (_ decode) Float(buf []byte) float32 {
	// Extract the uint32 value from the byte slice
	bits := binary.BigEndian.Uint32(buf)

	// Convert the uint32 value to a float32 value
	value := math.Float32frombits(bits)

	return value
}