package jsonport

import (
	"encoding/json"
	"fmt"
)

func Example_usage() {
	jsonstr := `{
		"timestamp": "1438194274",
		"users": [{"id": 1, "name": "Tom"}, {"id": 2, "name": "Peter"}],
		"keywords": ["golang", "json"]
	}`
	j, _ := Unmarshal([]byte(jsonstr))

	// Output: Tom
	fmt.Println(j.GetString("users", 0, "name"))

	// Output: [golang json]
	fmt.Println(j.Get("keywords").StringArray())

	// Output: [Tom Peter]
	names, _ := j.Get("users").EachOf("name").StringArray()
	fmt.Println(names)

	// try parse STRING as NUMBER
	j.StringAsNumber()
	// Output: 1438194274
	fmt.Println(j.Get("timestamp").Int())

	// convert NUMBER, STRING, ARRAY and OBJECT type to BOOL
	j.AllAsBool()
	// Output: true
	fmt.Println(j.GetBool("status"))
	// Output:
	// Tom <nil>
	// [golang json] <nil>
	// [Tom Peter]
	// 1438194274 <nil>
	// false <nil>
}

func ExampleJson_StringAsNumber() {
	jsonstr := `{"timestamp": "1438194274"}`
	j, _ := Unmarshal([]byte(jsonstr))

	fmt.Println("Without StringAsNumber():")
	n, err := j.GetInt("timestamp")
	fmt.Println(n, err)

	fmt.Println("With StringAsNumber():")
	j.StringAsNumber()
	n, err = j.GetInt("timestamp")
	fmt.Println(n, err)
	// Output:
	// Without StringAsNumber():
	// 0 type mismatch: expected NUMBER, found STRING "1438194274"
	// With StringAsNumber():
	// 1438194274 <nil>
}

func ExampleJson_AllAsBool() {
	jsonstr := `{"enabled": 1}`
	j, _ := Unmarshal([]byte(jsonstr))

	fmt.Println("Without AllAsBool():")
	b, err := j.GetBool("enabled")
	fmt.Println(b, err)

	fmt.Println("With AllAsBool():")
	j.AllAsBool()
	b, err = j.GetBool("enabled")
	fmt.Println(b, err)
	// Output:
	// Without AllAsBool():
	// false type mismatch: expected BOOL, found NUMBER "1"
	// With AllAsBool():
	// true <nil>

}

func ExampleNewJson() {
	jsonstr := `{"key": "score", "value": 99.5}`
	i := struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	}{}
	// using encoding/json
	json.Unmarshal([]byte(jsonstr), &i)
	j := NewJson(i.Value)
	// Output: score
	fmt.Println(i.Key)
	// Output: NUMBER 99.5
	f, _ := j.Float()
	fmt.Println(j.Type(), f)

	jsonstr = `{"key": "name", "value": "Tom"}`
	json.Unmarshal([]byte(jsonstr), &i)
	j = NewJson(i.Value)
	// Output: name
	fmt.Println(i.Key)
	// Output: STRING Tom
	s, _ := j.String()
	fmt.Println(j.Type(), s)
	// Output:
	// score
	// NUMBER 99.5
	// name
	// STRING Tom
}
