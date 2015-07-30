package jsonport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

const (
	zero = json.Number("0")
)

// Json represents everything of json
type Json struct {
	v    interface{}
	atoi bool
	tob  bool
	err  error
}

// Unmarshal parses data to Json
func Unmarshal(data []byte) (Json, error) {
	r := bytes.NewBuffer(data)
	return DecodeFrom(r)
}

// DecodeFrom parses data from reader to json
func DecodeFrom(r io.Reader) (Json, error) {
	var j Json
	dec := json.NewDecoder(r)
	// we use json.Number to prevent from int64 precision loss
	dec.UseNumber()
	err := dec.Decode(&j.v)
	if err != nil {
		return Json{}, err
	}
	return j, nil
}

// NewJson news Json with v.
// which *MUST* be decoded from `encoding/json` with interface{}.
func NewJson(v interface{}) Json {
	return Json{v: v}
}

// Value returns interface{} of current json value
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

// Type returns the Type of current json value
func (j Json) Type() Type {
	if j.err != nil {
		return INVALID
	}
	if j.v == nil {
		return NULL
	}
	switch j.v.(type) {
	default:
		return INVALID
	case bool:
		return BOOL
	case string:
		return STRING
	case []interface{}:
		return ARRAY
	case map[string]interface{}:
		return OBJECT
	case json.Number:
		return NUMBER
	case float64:
		// packakge 'encoding/json' convert number to float64 by default
		return NUMBER
	}
}

func (j Json) mismatch(t Type) error {
	if j.v == nil && j.err != nil {
		return j.err
	}
	return fmt.Errorf("type mismatch: expected %s, found %s %#v", t, j.Type(), j.v)
}

// String converts current json value to string
func (j Json) String() (string, error) {
	s, ok := j.v.(string)
	if ok {
		return s, nil
	}
	return "", j.mismatch(STRING)
}

func (j Json) number() (json.Number, error) {
	if n, ok := j.v.(json.Number); ok {
		return n, nil
	}
	if !j.atoi {
		return zero, j.mismatch(NUMBER)
	}
	if s, ok := j.v.(string); ok {
		return (json.Number)(s), nil
	}
	return zero, j.mismatch(NUMBER)
}

func tryInt64(n json.Number) (int64, error) {
	intn, err := n.Int64()
	if err == nil {
		return intn, nil
	}
	// fallback. convert float to int
	floatn, err1 := n.Float64()
	if err1 != nil {
		return 0, err1
	}
	uintn := uint64(floatn)
	if floatn != 0 && uintn == 0 {
		return 0, err
	}
	return int64(uintn), nil
}

// Float converts current json value to float64
func (j Json) Float() (float64, error) {
	if f, ok := j.v.(float64); ok {
		return f, nil
	}
	n, err := j.number()
	if err != nil {
		return 0, err
	}
	return n.Float64()
}

// Int converts current json value to int64
// result will be negative if pasring from uint64 like `1<<63`
func (j Json) Int() (int64, error) {
	if f, ok := j.v.(float64); ok {
		intn := uint64(f)
		if f != 0 && intn == 0 {
			return 0, fmt.Errorf("float %f overflows int64", f)
		}
		return int64(intn), nil
	}
	n, err := j.number()
	if err != nil {
		return 0, err
	}
	return tryInt64(n)
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
	case []interface{}:
		return len(t), nil
	case map[string]interface{}:
		return len(t), nil
	}
	return 0, fmt.Errorf("type %s not supported Len()", j.Type())
}

// Keys returns the field names of json object.
// error is returned if value type not equal to OBJECT.
func (j Json) Keys() ([]string, error) {
	m, ok := j.v.(map[string]interface{})
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
	m, ok := j.v.(map[string]interface{})
	if !ok {
		return nil, j.mismatch(OBJECT)
	}
	ret := make([]Json, 0, len(m))
	for _, v := range m {
		ret = append(ret, Json{v: v})
	}
	return ret, nil
}

// Member returns the member value specified by `name`
// a NULL type Json is returned if member not found
// Json.Error() is set if type not equal to OBJECT
func (j Json) Member(name string) Json {
	m, ok := j.v.(map[string]interface{})
	if !ok {
		return Json{err: j.mismatch(OBJECT)}
	}
	j.v, _ = m[name]
	// Json with nil value, that's fine
	return j
}

// Element returns the (i+1)th element of array.
// a NULL type Json is returned if index out of range.
// Json.Error() is set if type not equal to ARRAY.
func (j Json) Element(i int) Json {
	arr, ok := j.v.([]interface{})
	if !ok {
		return Json{err: j.mismatch(ARRAY)}
	}
	if i < 0 || len(arr)-1 < i {
		return Json{}
	}
	j.v = arr[i]
	return j
}

// Array converts current json value to []Json.
// error is returned if value type not equal to ARRAY.
func (j Json) Array() ([]Json, error) {
	arr, ok := j.v.([]interface{})
	if !ok {
		return nil, j.mismatch(ARRAY)
	}
	ret := make([]Json, len(arr))
	for i, e := range arr {
		ret[i].v = e
	}
	return ret, nil
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
	retv := make([]interface{}, len(arr))
	for i, e := range arr {
		j := e.Get(keys...)
		if j.err != nil {
			return Json{err: j.err}
		}
		retv[i] = j.v
	}
	return Json{v: retv}
}
