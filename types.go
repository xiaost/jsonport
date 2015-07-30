package jsonport

import (
	"fmt"
)

type Type int

const (
	INVALID Type = iota
	OBJECT
	ARRAY
	STRING
	NUMBER
	BOOL
	NULL
)

func (t Type) String() string {
	switch t {
	case INVALID:
		return "INVALID"
	case OBJECT:
		return "OBJECT"
	case ARRAY:
		return "ARRAY"
	case STRING:
		return "STRING"
	case NUMBER:
		return "NUMBER"
	case BOOL:
		return "BOOL"
	case NULL:
		return "NULL"
	}
	return fmt.Sprintf("UNKNOWN(%d)", t)
}
