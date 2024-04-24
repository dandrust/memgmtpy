// Tuple immplementation
package tuple

import (
	"encoding/binary"
	"memgmtgo/datatype"
	"memgmtgo/schema"
	"fmt"
)

const (
	lengthSize            = 2
	offsetArrayLengthSize = 2
	offsetPointerSize     = 2
)

func Encode(values []interface{}, schema schema.Schema) []byte {
	// Here we know that we'll have a 2 byte tuple
	// length, a 2 byte asset length, and one 2 byte
	// offset per asset.  Create a slice of that
	// capaity, and then we'll append assets onto that
	// as we go
	headerLength := lengthSize + offsetArrayLengthSize + (offsetPointerSize * schema.AssetCount)
	bytes := make([]byte, headerLength)

	// Write the number of slots!
	binary.BigEndian.PutUint16(bytes[lengthSize:], uint16(schema.AssetCount))

	// We'll start writing offsets after two x 2 byte ints
	// and assets will be appended to the end
	assetOffsetPos := lengthSize + offsetArrayLengthSize
	assetPos := len(bytes)

	for i, value := range values {
		field := schema.Fields[i]
		var to_write []byte

		switch field.DataType {
		case datatype.String:
			v, ok := value.(string)
			if !ok {
				panic("expected string!")
			}
			to_write = datatype.Encode.String(v)
		case datatype.Integer:
			v, ok := value.(int32)
			if !ok {
				panic("expected integer!")
			}

			to_write = datatype.Encode.Integer(v)
		case datatype.BigInteger:
			v, ok := value.(int64)
			if !ok {
				panic("expected big integer!")
			}

			to_write = datatype.Encode.BigInteger(v)
		case datatype.Boolean:
			v, ok := value.(bool)
			if !ok {
				panic("expected boolean!")
			}

			to_write = datatype.Encode.Boolean(v)
		case datatype.Float:
			v, ok := value.(float32)
			if !ok {
				panic("expected float!")
			}

			to_write = datatype.Encode.Float(v)
		}

		// Write the assetPos value at assetOffsetPos
		binary.BigEndian.PutUint16(bytes[assetOffsetPos:], uint16(assetPos))
		assetOffsetPos += offsetPointerSize

		// Append the bytes
		bytes = append(bytes, to_write...)
		assetPos += len(to_write)
	}

	// Write the tuple length now that we know it
	binary.BigEndian.PutUint16(bytes[:], uint16(len(bytes)))

	return bytes
}

func Decode(bytes[]byte, schema schema.Schema) {
	// In this context we don't care about tuple length, but it
	// will be helpful when we eventually need to pull arbitrary
	// tuples from a page

	assetCount := int(binary.BigEndian.Uint16(bytes[lengthSize:]))
	fmt.Println("there are", assetCount, "assets")

	offsetEntryPos := lengthSize + offsetArrayLengthSize
	fmt.Println("asset entries start at pos", offsetEntryPos)
	

	for i := 0; i < assetCount; i++ {
		// read the offset
		offset := binary.BigEndian.Uint16(bytes[offsetEntryPos:])
		fmt.Println("The offset at entry", i, "is", offset)
		offsetEntryPos += offsetPointerSize

		switch schema.Fields[i].DataType {
		case datatype.Integer:
			fmt.Println(schema.Fields[i].Name, ":", datatype.Decode.Integer(bytes[offset:]))
		}

		
	}
}
