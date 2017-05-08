package jsonport

import "errors"

var (
	errArrayIndex = errors.New("not array index")
	errMemberName = errors.New("not member name")

	errJSONEOF   = errors.New("JSON: unexpect EOF")
	errArrayEOF  = errors.New("ARRAY: unexpect EOF")
	errObjectEOF = errors.New("OBJECT: unexpect EOF")
	errStringEOF = errors.New("STRING: unexpect EOF")
)
