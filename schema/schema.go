package schema

type Field struct {
	Name                string
	DataType            int
	AssetNumber         int
	Nullable            bool
}

type Schema struct {
	Name                 string
	AssetCount           int
	Fields               []Field
}

func NewSchema(name string) *Schema {
	s := Schema{Name: name, AssetCount: 0, Fields: []Field{}}
	return &s
}

func (s *Schema) Add(dt int, name string, nullable bool) {
	f := Field{Name: name, DataType: dt, Nullable: nullable}

	f.AssetNumber = s.AssetCount
	s.AssetCount += 1

	s.Fields = append(s.Fields, f)
}
