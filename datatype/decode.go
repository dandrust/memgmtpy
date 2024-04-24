package datatype

import (
	"encoding/binary"
	// "math"
)

type decode struct{}


func (_ decode) Integer(buf []byte) int32 {
	return int32(binary.BigEndian.Uint32(buf))
}


func (_ decode) String(_ []byte) string {
	return "TBI"
}