package jsonport

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
)

const (
	zero = Number("0")
)

// Json represents everything of json
type Json struct {
	atoi bool
	tob  bool
	err  error

	/*
		v is one of:
			map[string]Json (tp: OBJECT)
			[]Json			(tp: ARRAY)
			Number			(tp: NUMBER)
			string			(tp: STRING)
			bool			(tp: BOOL)
			nil				(tp: NULL)
	*/
	v  interface{}
	tp Type
}

var (
	ErrMoreBytes = errors.New("not all bytes unmarshaled")
)

// Unmarshal parses data to Json
func Unmarshal(data []byte) (Json, error) {
	j, i, err := parse(data)
	if err != nil {
		return j, err
	}
	n := skipSpace(data[i:])
	if n+i != len(data) {
		return j, ErrMoreBytes
	}
	return j, nil
}

// DecodeFrom parses data from reader to json
func DecodeFrom(r io.Reader) (Json, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return Json{}, nil
	}
	return Unmarshal(b)
}

// Type returns the Type of current json value
func (j Json) Type() Type {
	return j.tp
}

/* Value returns interface{} of current json value
Value is one of:
	map[string]Json (tp: OBJECT)
	Number			(tp: NUMBER)
	[]Json			(tp: ARRAY)
	string			(tp: STRING | NUMBER)
	bool			(tp: BOOL)
	nil				(tp: NULL)
*/
func (j Json) Value() interface{} {
	return j.v
}

// Error returns error of current context
func (j Json) Error() error {
	return j.err
}

// StringAsNumber enables conversion like {"id": "123"},
// j.GetInt("id") is returned 123 instead of an err,
// j.GetString("id") is returned "123" as expected.
func (j *Json) StringAsNumber() {
	j.atoi = true
}

// AllAsBool enables conversion of Bool():
//	STRING, ARRAY, OBJECT:	true if Len() != 0
//	NUMBER:	true if Float() != 0
//	NULL:	false
func (j *Json) AllAsBool() {
	j.tob = true
}

func (j Json) mismatch(t Type) error {
	if j.err != nil {
		return j.err
	}
	return fmt.Errorf("type mismatch: expected %s, found %s", t, j.tp)
}

// String converts current json value to string
func (j Json) String() (string, error) {
	s, ok := j.v.(string)
	if ok {
		return s, nil
	}
	return "", j.mismatch(STRING)
}

func (j Json) number() (Number, error) {
	if n, ok := j.v.(Number); ok {
		return n, nil
	}
	if !j.atoi {
		return zero, j.mismatch(NUMBER)
	}
	if s, ok := j.v.(string); ok {
		return Number(s), nil
	}
	return zero, j.mismatch(NUMBER)
}

// Float converts current json value to float64
func (j Json) Float() (float64, error) {
	n, err := j.number()
	if err != nil {
		return 0, err
	}
	return n.Float64()
}

// Int converts current json value to int64
// result will be negative if pasring from uint64 like `1<<63`
func (j Json) Int() (int64, error) {
	n, err := j.number()
	if err != nil {
		return 0, err
	}
	i, err := n.Int64()
	if err != nil {
		if f, err2 := n.Float64(); err2 == nil {
			if f < 9.223372036e+21 && f > -9.223372036e+21 {
				return int64(f), nil
			}
		}
	}
	return i, err
}

// Bool converts current json value to bool
func (j Json) Bool() (bool, error) {
	b, ok := j.v.(bool)
	if ok {
		return b, nil
	}
	if !j.tob {
		return false, j.mismatch(BOOL)
	}
	switch j.Type() {
	case STRING, ARRAY, OBJECT:
		l, err := j.Len()
		return l > 0, err
	case NUMBER:
		f, err := j.Float()
		return f != 0, err
	case NULL:
		return false, nil
	}
	return false, j.mismatch(BOOL)
}

// Len returns the length of json value.
//	STRING:	the number of bytes
//	ARRAY:	the number of elements
//	OBJECT:	the number of pairs
func (j Json) Len() (int, error) {
	switch t := j.v.(type) {
	case string:
		return len(t), nil
	case []Json:
		return len(t), nil
	case map[string]Json:
		return len(t), nil
	}
	return 0, fmt.Errorf("type %s not supported Len()", j.Type())
}

// Keys returns the field names of json object.
// error is returned if value type not equal to OBJECT.
func (j Json) Keys() ([]string, error) {
	m, ok := j.v.(map[string]Json)
	if !ok {
		return nil, j.mismatch(OBJECT)
	}
	ret := make([]string, 0, len(m))
	for k := range m {
		ret = append(ret, k)
	}
	return ret, nil
}

// Values returns the field values of json object.
// error is returned if value type not equal to OBJECT.
func (j Json) Values() ([]Json, error) {
	m, ok := j.v.(map[string]Json)
	if !ok {
		return nil, j.mismatch(OBJECT)
	}
	ret := make([]Json, 0, len(m))
	for _, v := range m {
		ret = append(ret, v)
	}
	return ret, nil
}

// Member returns the member value specified by `name`
// a NULL type Json is returned if member not found
// Json.Error() is set if type not equal to OBJECT
func (j Json) Member(name string) Json {
	m, ok := j.v.(map[string]Json)
	if !ok {
		return Json{err: j.mismatch(OBJECT)}
	}
	v, ok := m[name]
	if !ok {
		j.v = nil
		j.tp = NULL
	} else {
		j.v = v.v
		j.tp = v.tp
	}
	return j
}

// Element returns the (i+1)th element of array.
// a NULL type Json is returned if index out of range.
// Json.Error() is set if type not equal to ARRAY.
func (j Json) Element(i int) Json {
	arr, ok := j.v.([]Json)
	if !ok {
		return Json{err: j.mismatch(ARRAY)}
	}
	if i < 0 || i >= len(arr) {
		j.v = nil
		j.tp = NULL
	} else {
		j.v = arr[i].v
		j.tp = arr[i].tp
	}
	return j
}

// Array converts current json value to []Json.
// error is returned if value type not equal to ARRAY.
func (j Json) Array() ([]Json, error) {
	arr, ok := j.v.([]Json)
	if !ok {
		return nil, j.mismatch(ARRAY)
	}
	return arr, nil
}

// IntArray converts current json value to []int64.
// error is returned if any element type not equal to NUMBER.
func (j Json) IntArray() ([]int64, error) {
	arr, err := j.Array()
	if err != nil {
		return nil, err
	}
	ret := make([]int64, len(arr))
	for i, e := range arr {
		n, err := e.Int()
		if err != nil {
			return nil, err
		}
		ret[i] = n
	}
	return ret, nil
}

// FloatArray converts current json value to []float64.
// error is returned if any element type not equal to NUMBER.
func (j Json) FloatArray() ([]float64, error) {
	arr, err := j.Array()
	if err != nil {
		return nil, err
	}
	ret := make([]float64, len(arr))
	for i, e := range arr {
		n, err := e.Float()
		if err != nil {
			return nil, err
		}
		ret[i] = n
	}
	return ret, nil
}

// BoolArray converts current json value to []bool.
// error is returned if any element type not equal to NUMBER.
func (j Json) BoolArray() ([]bool, error) {
	arr, err := j.Array()
	if err != nil {
		return nil, err
	}
	ret := make([]bool, len(arr))
	for i, e := range arr {
		ret[i], err = e.Bool()
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

// StringArray converts current json value to []string.
// error is returned if any element type not equal to STRING.
func (j Json) StringArray() ([]string, error) {
	arr, err := j.Array()
	if err != nil {
		return nil, err
	}
	ret := make([]string, len(arr))
	for i, e := range arr {
		ret[i], err = e.String()
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

// Get returns Json object by key sequence.
//	key with the type of string is equal to j.Member(k),
//	key with the type of number is equal to j.Element(k),
//	j.Get("key", 1) is equal to j.Member("key").Element(1).
// a NULL type Json returned with err if:
//	- key type not supported. (neither number nor string)
//	- json value type mismatch.
func (j Json) Get(keys ...interface{}) Json {
	for _, k := range keys {
		if j.err != nil {
			return j
		}
		switch t := k.(type) {
		// reflect.ValueOf(t).Int() or Uint() ?
		// without reflection here.
		case int:
			j = j.Element(t)
		case int8:
			j = j.Element(int(t))
		case int16:
			j = j.Element(int(t))
		case int32:
			j = j.Element(int(t))
		case int64:
			j = j.Element(int(t))
		case uint:
			j = j.Element(int(t))
		case uint8:
			j = j.Element(int(t))
		case uint16:
			j = j.Element(int(t))
		case uint32:
			j = j.Element(int(t))
		case uint64:
			j = j.Element(int(t))
		case string:
			j = j.Member(t)
		default:
			return Json{err: fmt.Errorf("key type %T not supported", t)}
		}
	}
	return j
}

// GetBool convert json value specified by keys to bool,
// it is equal to Get(keys...).Bool()
func (j Json) GetBool(keys ...interface{}) (bool, error) {
	return j.Get(keys...).Bool()
}

// GetString convert json value specified by keys to string,
// it is equal to Get(keys...).String()
func (j Json) GetString(keys ...interface{}) (string, error) {
	return j.Get(keys...).String()
}

// GetFloat  convert json value specified by keys to float64,
// it is equal to Get(keys...).Float()
func (j Json) GetFloat(keys ...interface{}) (float64, error) {
	return j.Get(keys...).Float()
}

// GetInt convert json value specified by keys to int64,
// it is equal to Get(keys...).Int()
func (j Json) GetInt(keys ...interface{}) (int64, error) {
	return j.Get(keys...).Int()
}

// EachOf convert every elements specified by keys in json value to ARRAY.
// Json.Error() is set if an error occurred.
// it is equal to `Json{[e.Get(keys...) for e in j.Array()]}`
func (j Json) EachOf(keys ...interface{}) Json {
	var arr []Json
	var err error
	t := j.Type()
	if t == ARRAY {
		arr, err = j.Array()
	} else if t == OBJECT {
		arr, err = j.Values()
	} else {
		err = fmt.Errorf("type %s not supported EachOf()", j.Type())
	}
	if err != nil {
		return Json{err: err}
	}
	retv := make([]Json, len(arr))
	for i, e := range arr {
		j := e.Get(keys...)
		if j.err != nil {
			return Json{err: j.err}
		}
		retv[i] = j
	}
	return Json{v: retv, tp: ARRAY}
}
