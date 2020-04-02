
jsonport [![GoDoc](https://godoc.org/github.com/xiaost/jsonport?status.png)](https://godoc.org/github.com/xiaost/jsonport)
====

jsonport is a simple and high performance golang package for accessing json without pain. features:
* No reflection.
* Unmarshal without struct.
* Unmarshal for the given json path only.
* 2x faster than encoding/json.

It is inspired by [jmoiron/jsonq](https://github.com/jmoiron/jsonq). Feel free to post issues or PRs, I will reply ASAP :-)


## Usage

```go
package main

import (
	"fmt"

	"github.com/xiaost/jsonport"
)

func main() {
	jsonstr := `{
		"timestamp": "1438194274",
		"users": [{"id": 1, "name": "Tom"}, {"id": 2, "name": "Peter"}],
		"keywords": ["golang", "json"],
		"status": 1
	}`
	j, _ := jsonport.Unmarshal([]byte(jsonstr))

	fmt.Println(j.GetString("users", 0, "name"))             // Tom, nil
	fmt.Println(j.Get("keywords").StringArray())             // [golang json], nil
	fmt.Println(j.Get("users").EachOf("name").StringArray()) // [Tom Peter],  nil

	// try parse STRING as NUMBER
	fmt.Println(j.Get("timestamp").Int()) // 0, type mismatch: expected NUMBER, found STRING
	j.StringAsNumber()
	fmt.Println(j.Get("timestamp").Int()) // 1438194274, nil

	// convert NUMBER, STRING, ARRAY and OBJECT type to BOOL
	fmt.Println(j.GetBool("status")) // false, type mismatch: expected BOOL, found NUMBER
	j.AllAsBool()
	fmt.Println(j.GetBool("status")) // true, nil

	// using Unmarshal with path which can speed up json decode
	j, _ = jsonport.Unmarshal([]byte(jsonstr), "users", 1, "name")
	fmt.Println(j.String()) // Peter, nil
}
```

For more information on getting started with `jsonport` [check out the doc](https://godoc.org/github.com/xiaost/jsonport)
