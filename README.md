
jsonport [![GoDoc](https://godoc.org/github.com/xiaost/jsonport?status.png)](https://godoc.org/github.com/xiaost/jsonport)
====

jsonport is a golang package for accessing json with interface{} (schemaless) inspired by [jmoiron/jsonq](https://github.com/jmoiron/jsonq). Feel free to post issues or PRs, I will reply ASAP :-)


## Usage

```go
package main

import (
    "fmt"
    "github.com/xiaost/jsonport"
)

func main() {
    jsonstr := `{
        "status": 1,
        "timestamp": "1438194274",
        "users": [{"id": 1, "name": "Tom"}, {"id": 2, "name": "Peter"}],
        "keywords": ["golang", "json"]
    }`
    j, _ := jsonport.Unmarshal([]byte(jsonstr))

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
}


```

For more information on getting started with `jsonport` [check out the doc](https://godoc.org/github.com/xiaost/jsonport)

## Unmarshal Tricks

jsonport uses package `encoding/json` to unmarshal bytes. you can do it youself:

```go
jsonstr := `{"key": "score", "value": 99.5}`
i := struct {
    Key   string      `json:"key"`
    Value interface{} `json:"value"` // must be interface{}
}{}
// using encoding/json
json.Unmarshal([]byte(jsonstr), &i) 
j := jsonport.NewJson(i.Value)
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
```

## Performance
Er... jsonport is focused on simplicity, it working well with small json bytes. We recommend that you only define the few fields of struct with `interface{}` which may be variety. 
