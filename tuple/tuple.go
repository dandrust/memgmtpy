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

func writeNull(bytes []byte, assetOffsetPos int, nullBitFieldStart int, fieldNumber int) {
	// Calculate the null bitfield position
	p := nullBitFieldStart + (fieldNumber / 8)
	// Calclate the shift/offset
	shift := fieldNumber % 8
	fmt.Println("i is ", fieldNumber, ". p is", p, "and shift is", shift)
	// Set the appropriate null bit
	mask := byte(1 << shift)
	fmt.Printf("mask is %b\n", mask)
	// Write a zero pointer to the offset array
	bytes[p] ^= mask
	fmt.Printf("bitfield is now %b\n", bytes[p])
	
	binary.BigEndian.PutUint16(bytes[assetOffsetPos:], uint16(0))
}

func write(bytes []byte, assetOffsetPos int, assetPos int, fieldNumber int, to_write []byte) []byte {
	// Write the assetPos value at assetOffsetPos
	binary.BigEndian.PutUint16(bytes[assetOffsetPos:], uint16(assetPos))

	// Append the bytes
	return append(bytes, to_write...)
}

func Encode(values []interface{}, schema schema.Schema) []byte {
	// How many nullable fields do we have?
	var nullableCount int
	for _, f := range schema.Fields {
		if f.Nullable { nullableCount += 1 }
	}
	fmt.Println(schema.Name, "has", nullableCount, "nullable fields")

	// How many null bitfields do we need?
	nullBitFieldCount := nullableCount / 8
	if (nullableCount % 8) > 0 {
		nullBitFieldCount += 1
	}
	fmt.Println("We need", nullBitFieldCount, "bit fields")

	// Here we know that we'll have a 2 byte tuple
	// length, a 2 byte asset length, and one 2 byte
	// offset per asset, plus any null bitfields.  
	// Create a slice of that capaity, and then we'll 
	// append assets onto that as we go
	headerLength := lengthSize + offsetArrayLengthSize + (offsetPointerSize * schema.AssetCount) + nullBitFieldCount
	bytes := make([]byte, headerLength)

	// Populate the bit field arrays with 1's matching the number of fields we'll write. That way, any new fields
	// will be null by default
	nullBitFieldStart := headerLength - nullBitFieldCount
	j := 0
	for i := schema.AssetCount; i > 0; i -= 8 {
		var v byte
		if i < 8 {
			v = 1 << i
			v -= 1
		} else { 
			v = byte(0xFF)
		}
		fmt.Printf("Writing %b at pos %d\n", v, j)
		bytes[nullBitFieldStart + j] = v
		j += 1
	}

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
			switch v := value.(type) {
			case string:
				to_write = datatype.Encode.String(v)
				bytes = write(bytes, assetOffsetPos, assetPos, i, to_write)
				assetPos += len(to_write)
			case datatype.Null:
				if !field.Nullable { panic("encountered null for non nullable field!") }				

				writeNull(bytes, assetOffsetPos, nullBitFieldStart, i)
			}
		case datatype.Integer:
			v, ok := value.(int32)
			if !ok {
				panic("expected integer!")
			}

			to_write = datatype.Encode.Integer(v)
			bytes = write(bytes, assetOffsetPos, assetPos, i, to_write)
			assetPos += len(to_write)
		case datatype.BigInteger:
			v, ok := value.(int64)
			if !ok {
				panic("expected big integer!")
			}

			to_write = datatype.Encode.BigInteger(v)
			bytes = write(bytes, assetOffsetPos, assetPos, i, to_write)
			assetPos += len(to_write)
		case datatype.Boolean:
			v, ok := value.(bool)
			if !ok {
				panic("expected boolean!")
			}

			to_write = datatype.Encode.Boolean(v)
			bytes = write(bytes, assetOffsetPos, assetPos, i, to_write)
			assetPos += len(to_write)
		case datatype.Float:
			v, ok := value.(float32)
			if !ok {
				panic("expected float!")
			}

			to_write = datatype.Encode.Float(v)
			bytes = write(bytes, assetOffsetPos, assetPos, i, to_write)
			assetPos += len(to_write)
		}

		assetOffsetPos += offsetPointerSize
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

	offsetEntryPos := lengthSize + offsetArrayLengthSize

	nullBitFieldStart := offsetEntryPos + (offsetPointerSize * assetCount)

	for i := 0; i < assetCount; i++ {
		// is it null?
		p := i / 8
		mask := byte(1 << (i % 8))

		if (bytes[nullBitFieldStart + p] & mask) == 0 {
			fmt.Println(schema.Fields[i].Name, ": NULL")
			offsetEntryPos += offsetPointerSize
			continue
		}

		// read the offset
		offset := binary.BigEndian.Uint16(bytes[offsetEntryPos:])
		offsetEntryPos += offsetPointerSize

		switch schema.Fields[i].DataType {
		case datatype.Integer:
			fmt.Println(schema.Fields[i].Name, ":", datatype.Decode.Integer(bytes[offset:]))
		case datatype.BigInteger:
			fmt.Println(schema.Fields[i].Name, ":", datatype.Decode.BigInteger(bytes[offset:]))
		case datatype.Boolean:
			fmt.Println(schema.Fields[i].Name, ":", datatype.Decode.Boolean(bytes[offset:]))
		case datatype.Float:
			fmt.Println(schema.Fields[i].Name, ":", datatype.Decode.Float(bytes[offset:]))
		case datatype.String:
			fmt.Println(schema.Fields[i].Name, ":", datatype.Decode.String(bytes[offset:]))
		}
	}
}
