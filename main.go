package main

import (
	"fmt"
	"memgmtgo/datatype"
	"memgmtgo/schema"
	"memgmtgo/tuple"
	// "memgmtgo/page"
	// "os"
	// "encoding/binary"
)

func main() {
	personSchema := schema.NewSchema("people")

	personSchema.Add(datatype.BigInteger, "id", false)
	personSchema.Add(datatype.String, "name", false)
	personSchema.Add(datatype.Integer, "age", false)
	personSchema.Add(datatype.Boolean, "in_michigan", false)
	personSchema.Add(datatype.String, "nickname", true)
	personSchema.Add(datatype.Float, "temperature", true)

	values := []interface{}{
		int64(1),  		// id
		"Daniel",  		// name
		int32(36), 		// age
		true,      		// inMichigan
		"Dan",     		// nickname (nullable)
		float32(98.6),	// temperature (nullable)
	}

	bytes := tuple.Encode(values, *personSchema)

	fmt.Println(bytes)

	//  0                       1                      2                         3                         4
	//  0 1  2 3 4 5  6 7  8 9  0 1  2 3  4 5  6 7 8 9 0 1 2 3 4 5 6  7  8   9   0   1   2 3 4 5  6 7 8 9  0  1   2  3   4  5
	//
	// [0 46 0 6 0 16 0 24 0 32 0 36 0 37 0 42 0 0 0 0 0 0 0 1 0 6 68 97 110 105 101 108 0 0 0 36 1 0 3 68 97 110 66 197 51 51]
	//  ^    ^   ^    ^    ^    ^    ^    ^    ^               ^                         ^        ^ ^             ^
	//  |    |   |    |    |    |    |    |    + id			   + name                    + age    | + nickname    + temperature
	//  |    |   |    |    |    |    |    + temperature offset									  + inMichigan
	//  |    |   |    |    |    |    + nickname offset
	//  |    |   |    |    |    + inMichigan offset
	//  |    |   |    |    + age offset
	//  |    |   |    + name offset
	//  |    |   + id offset
	//  |    + asset count
	//  + tuple length

	tuple.Decode(bytes, *personSchema)
}
