package datatype

const (
	Integer = iota
	String
	Boolean
	BigInteger
	Float
)

type Null struct {}

var Encode = encode{}
var Decode = decode{}
