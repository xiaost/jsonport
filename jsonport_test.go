package jsonport

import (
	"reflect"
	"sort"
	"testing"
)

func TestJson_Number(t *testing.T) {
	// case simple interger
	in := []byte(`10000`)
	j, _ := Unmarshal(in)
	if j.Type() != NUMBER {
		t.Fatal(j.Type())
	}
	if intn, err := j.Int(); intn != int64(10000) || err != nil {
		t.Fatal(intn, err)
	}
	if floatn, _ := j.Float(); floatn != float64(10000) {
		t.Fatal(floatn)
	}
	// case simple float
	in = []byte(`1024.125`)
	j, _ = Unmarshal(in)
	if j.Type() != NUMBER {
		t.Fatal(j.Type())
	}
	if intn, err := j.Int(); intn != int64(1024) {
		t.Fatal(intn, err)
	}
	if floatn, err := j.Float(); floatn != float64(1024.125) {
		t.Fatal(j.Value(), err)
	}
	// case overflow int64
	in = []byte(`9223372036000000000000`)
	j, _ = Unmarshal(in)
	if j.Type() != NUMBER {
		t.Fatal(j.Type())
	}
	if n, err := j.Int(); err == nil {
		t.Fatal(n, err)
	}
	floatn, _ := j.Float()
	expfloat := float64(9.223372036e+21)
	if floatn != expfloat {
		t.Fatal(floatn, expfloat)
	}
	// case use StringAsNumber
	in = []byte(`"10000"`)
	j, _ = Unmarshal(in)
	if j.Type() != STRING {
		t.Fatal(j.Type())
	}
	j.StringAsNumber()
	if intn, _ := j.Int(); intn != int64(10000) {
		t.Fatal(intn)
	}
	if floatn, _ := j.Float(); floatn != float64(10000) {
		t.Fatal(floatn)
	}
	// case not a number
	in = []byte(`"a10000"`)
	j, _ = Unmarshal(in)
	if j.Type() != STRING {
		t.Fatal(j.Type())
	}
	if _, err := j.Int(); err == nil {
		t.Fatal(err)
	}
	j.StringAsNumber()
	if _, err := j.Float(); err == nil {
		t.Fatal(err)
	}

	// case float overflow
	in = []byte(`1e9999`)
	j, _ = Unmarshal(in)
	if j.Type() != NUMBER {
		t.Fatal(j.Type())
	}
	if _, err := j.Int(); err == nil {
		t.Fatal(err)
	}
	if _, err := j.Float(); err == nil {
		t.Fatal(err)
	}

}

func TestJson_Bool(t *testing.T) {
	// case true
	in := []byte(`true`)
	j, _ := Unmarshal(in)
	if j.Type() != BOOL {
		t.Fatal(j.Type())
	}
	if b, err := j.Bool(); b != true || err != nil {
		t.Fatal(b, err)
	}
	// case false
	in = []byte(`false`)
	j, _ = Unmarshal(in)
	if j.Type() != BOOL {
		t.Fatal(j.Type())
	}
	if b, err := j.Bool(); b != false || err != nil {
		t.Fatal(b, err)
	}

	// case AllAsBool
	in = []byte(`"s"`)
	j, _ = Unmarshal(in)
	j.AllAsBool()
	if j.Type() != STRING {
		t.Fatal(j.Type())
	}
	if b, err := j.Bool(); b != true || err != nil {
		t.Fatal(b, err)
	}

	in = []byte(`0`)
	j, _ = Unmarshal(in)
	j.AllAsBool()
	if j.Type() != NUMBER {
		t.Fatal(j.Type())
	}
	if b, err := j.Bool(); b != false || err != nil {
		t.Fatal(b, err)
	}

	in = []byte(`null`)
	j, _ = Unmarshal(in)
	j.AllAsBool()
	if j.Type() != NULL {
		t.Fatal(j.Type())
	}
	if b, err := j.Bool(); b != false || err != nil {
		t.Fatal(b, err)
	}
}

func TestJson_String(t *testing.T) {
	// simple string
	in := `"blabla"`
	j, _ := Unmarshal([]byte(in))
	if j.Type() != STRING {
		t.Fatal(j.Type())
	}
	in = in[1 : len(in)-1]
	if n, err := j.Len(); err != nil || n != len(in) {
		t.Fatal(n, err)
	}
	if s, _ := j.String(); s != in {
		t.Fatal(s)
	}

	// case ""
	in = `""`
	j, _ = Unmarshal([]byte(in))
	if j.Type() != STRING {
		t.Fatal(j.Type())
	}
	in = in[1 : len(in)-1]
	if n, err := j.Len(); err != nil || n != len(in) {
		t.Fatal(n, err)
	}
	if s, _ := j.String(); s != in {
		t.Fatal(s)
	}
}

func TestJson_Null(t *testing.T) {
	in := `null`
	j, _ := Unmarshal([]byte(in))
	if j.Type() != NULL {
		t.Fatal(j.Type())
	}
	if _, err := j.Len(); err == nil {
		t.Fatal(err)
	}
}

func TestJson_Array(t *testing.T) {
	// case num array
	in := `[1, 2, 3]`
	j, _ := Unmarshal([]byte(in))
	if j.Type() != ARRAY {
		t.Fatal(j.Error())
	}
	if n, _ := j.Len(); n != 3 {
		t.Fatal(n)
	}
	narr, err := j.IntArray()
	if !reflect.DeepEqual(narr, []int64{1, 2, 3}) {
		t.Fatal(narr, err)
	}
	farr, _ := j.FloatArray()
	if !reflect.DeepEqual(farr, []float64{1, 2, 3}) {
		t.Fatal(narr)
	}

	// case num array overfloat
	in = `[1, 2e999, 3]`
	j, _ = Unmarshal([]byte(in))
	if j.Type() != ARRAY {
		t.Fatal(j.Type())
	}
	if narr, err := j.IntArray(); err == nil {
		t.Fatal(narr)
	}
	if narr, err := j.FloatArray(); err == nil {
		t.Fatal(narr)
	}
	// case bool array
	in = `[true, false, true]`
	j, _ = Unmarshal([]byte(in))
	if j.Type() != ARRAY {
		t.Fatal(j.Type())
	}
	bar, _ := j.BoolArray()
	if !reflect.DeepEqual(bar, []bool{true, false, true}) {
		t.Fatal(bar)
	}
	// case string array
	in = `["1", "2", "3"]`
	j, _ = Unmarshal([]byte(in))
	if j.Type() != ARRAY {
		t.Fatal(j.Type())
	}
	sar, _ := j.StringArray()
	if !reflect.DeepEqual(sar, []string{"1", "2", "3"}) {
		t.Fatal(sar)
	}
	// case mix array
	in = `[1, "2", true]`
	j, _ = Unmarshal([]byte(in))
	if j.Type() != ARRAY {
		t.Fatal(j.Type())
	}
	if n, _ := j.Element(0).Int(); n != 1 {
		t.Fatal(n)
	}
	if s, _ := j.Element(1).String(); s != "2" {
		t.Fatal(s)
	}
	if b, _ := j.Element(2).Bool(); b != true {
		t.Fatal(b)
	}

	if tt := j.Element(3).Type(); tt != NULL {
		t.Fatal(tt)
	}
}

func TestJson_Object(t *testing.T) {
	// case empty object
	in := `{}`
	j, _ := Unmarshal([]byte(in))
	if j.Type() != OBJECT {
		t.Fatal(j.Type())
	}
	if n, err := j.Len(); n != 0 || err != nil {
		t.Fatal(n, err)
	}
	if keys, err := j.Keys(); len(keys) != 0 || err != nil {
		t.Fatal(keys, err)
	}
	if values, err := j.Values(); len(values) != 0 || err != nil {
		t.Fatal(values, err)
	}

	// case simple object
	in = `{"k1":"v1"}`
	j, _ = Unmarshal([]byte(in))
	if j.Type() != OBJECT {
		t.Fatal(j.Error())
	}
	if n, err := j.Len(); n != 1 || err != nil {
		t.Fatal(n, err)
	}
	if keys, err := j.Keys(); len(keys) != 1 || keys[0] != "k1" || err != nil {
		t.Fatal(keys, err)
	}
	if values, err := j.Values(); len(values) != 1 || values[0].v.(string) != "v1" || err != nil {
		t.Fatal(values, err)
	}
	jj := j.Member("k1")
	if jj.Type() != STRING {
		t.Fatal(j.Type())
	}
	if s, _ := jj.String(); s != "v1" {
		t.Fatal(s)
	}
}

func TestJson_Get(t *testing.T) {
	// case simple
	in := `{"b": true, "num": 1, "array": ["2"]}`
	j, _ := Unmarshal([]byte(in))
	if j.Type() != OBJECT {
		t.Fatal(j.Type())
	}
	if b, err := j.GetBool("b"); b != true || err != nil {
		t.Fatal(b, err)
	}
	if n, err := j.GetInt("num"); n != 1 || err != nil {
		t.Fatal(n, err)
	}
	if n, err := j.GetFloat("num"); n != 1 || err != nil {
		t.Fatal(n, err)
	}
	if s, err := j.GetString("array", 0); s != "2" || err != nil {
		t.Fatal(s, err)
	}
	jj := j.Member("array")
	// case number key types
	keys := []interface{}{int(0), int8(0), int16(0), int32(0), int64(0)}
	keys = append(keys, []interface{}{uint(0), uint8(0), uint16(0), uint32(0), uint64(0)}...)
	for _, k := range keys {
		if s, err := jj.GetString(k); s != "2" || err != nil {
			t.Fatal(s, err)
		}
	}
}

func TestJson_EachOf(t *testing.T) {
	// case array
	in := `[{"id": 1}, {"id": 2}, {"id": 3}]`
	j, err := Unmarshal([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	j = j.EachOf("id")
	if j.Type() != ARRAY {
		t.Fatal(j.Type())
	}
	narr, err := j.IntArray()
	if !reflect.DeepEqual(narr, []int64{1, 2, 3}) {
		t.Fatal(narr, err)
	}
	// case object
	in = `{"a":{"name": "Tom"}, "b": {"name": "Peter"}, "c": {"name": "Mary"}}`
	j, _ = Unmarshal([]byte(in))
	j = j.EachOf("name")
	if j.Type() != ARRAY {
		t.Fatal(j.Type())
	}
	sarr, err := j.StringArray()
	sort.StringSlice(sarr).Sort()
	if !reflect.DeepEqual(sarr, []string{"Mary", "Peter", "Tom"}) {
		t.Fatal(sarr, err)
	}
	// case type err
	in = `123`
	j, _ = Unmarshal([]byte(in))
	j = j.EachOf()
	if j.Error() == nil {
		t.Fatal(nil)
	}
	// case key err
	in = `[{"id": 1}, {"id": 2}, {"id": 3}]`
	j, _ = Unmarshal([]byte(in))
	j = j.EachOf("name", 0)
	if j.Type() != INVALID {
		t.Fatal(j.Type())
	}
}

func TestJson_Mismatch(t *testing.T) {
	// case simple mismatch
	in := `123`
	j, _ := Unmarshal([]byte(in))
	if _, err := j.String(); err == nil {
		t.Fatal(nil)
	}
	if _, err := j.Bool(); err == nil {
		t.Fatal(nil)
	}
	if _, err := j.Keys(); err == nil {
		t.Fatal(nil)
	}
	if _, err := j.Values(); err == nil {
		t.Fatal(nil)
	}
	if jj := j.Member("1"); jj.Error() == nil {
		t.Fatal(nil)
	}
	if jj := j.Element(1); jj.Error() == nil {
		t.Fatal(nil)
	}
	if _, err := j.Array(); err == nil {
		t.Fatal(nil)
	}
	if _, err := j.FloatArray(); err == nil {
		t.Fatal(nil)
	}
	if _, err := j.IntArray(); err == nil {
		t.Fatal(nil)
	}
	if _, err := j.StringArray(); err == nil {
		t.Fatal(nil)
	}
	if _, err := j.BoolArray(); err == nil {
		t.Fatal(nil)
	}
	if jj := j.Get("x", "y"); jj.Error() == nil {
		t.Fatal(nil)
	}

	// array mismatch
	in = `[123, "1"]`
	j, _ = Unmarshal([]byte(in))
	if _, err := j.FloatArray(); err == nil {
		t.Fatal(nil)
	}
	if _, err := j.IntArray(); err == nil {
		t.Fatal(nil)
	}
	if _, err := j.StringArray(); err == nil {
		t.Fatal(nil)
	}
	if _, err := j.BoolArray(); err == nil {
		t.Fatal(nil)
	}

}

func TestJson_Other(t *testing.T) {
	in := `[123, "1"]`
	j, _ := Unmarshal([]byte(in))
	// case key type not support
	jj := j.Get(j)
	if jj.Error() == nil {
		t.Fatal(nil)
	}
	if jj.Value() != nil {
		t.Fatal(jj.Value())
	}
	if jj.Type() != INVALID {
		t.Fatal(jj.Type())
	}
	// case convert on err
	if _, err := jj.Float(); err == nil {
		t.Fatal(nil)
	}
	if jj.Error() == nil {
		t.Fatal(nil)
	}
	jj.AllAsBool()
	if _, err := jj.Bool(); err == nil {
		t.Fatal(nil)
	}
	// case decode err
	if _, err := Unmarshal([]byte("!")); err == nil {
		t.Fatal(nil)
	}

}

func TestTypes(t *testing.T) {
	tt := Type(255)
	if tt.String() != "UNKNOWN(255)" {
		t.Fatal(tt.String())
	}
}
