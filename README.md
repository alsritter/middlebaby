TODO: ...

use example see https://gitee.com/alsritter/testmb

config file.

```yml
port: 7689
httpFiles: 
  - ./tests/test.impl.json
name: username
watcher: true
cors:
  methods: ["GET"]
  headers: ["Content-Type"]
  exposed_headers: ["Cache-Control"]
  origins: ["*"]
  allow_credentials: true
```

http mock file

```json
[
  {
    "request": {
      "method": "GET",
      "url": "http://example.org/get",
      "params": {
        "name": "John",
        "age": "55"
      }
    },
    "response": {
      "status": 200,
      "headers": {
        "Content-Type": "application/json"
      },
      "delay": {
        "delay": 300,
        "offset": 200
      },
      "body": "{\"data\":{\"name\":\"John\",\"color\":\"Purples\",\"age\":55}}"
    }
  }
]
```

use example:

```go
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	resp, err := http.Get("http://example.org/get?name=John&age=55")
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response: ", string(body))

	select {} // blocking, test file modification
}
```

makefile:

```makefile
BIN_FILE=testmb

.PHONY: buildandrun
buildandrun:
	@go build -o "${BIN_FILE}" main.go
	middlebaby server --log-level="DEBUG" --app="./${BIN_FILE}"
	rm testmb
```