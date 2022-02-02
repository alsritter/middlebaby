# Middlebaby

Middlebaby is a tool for Mock services and interface testing of services under test

TODO: Perfect README

use example see https://gitee.com/alsritter/testmb

## Overview
Middlebaby is a tool for Mock services and interface testing of services under test

Middlebaby is committed to improving the work quality and efficiency of developers and testers, ensuring that the test quality reduces the test workload of service follow-up maintenance.With the help of tools, developers can get rid of the limitations of various dependencies in the self-testing stage, and quickly carry out interface availability verification and scene verification.Testers use tools to verify interface scenarios, and collaborate with developers to write and supplement interface test cases to improve collaboration efficiency and accuracy of information synchronization.

Middlebaby provides:
* Mock MySQL and Redis Storage.
* An easy way to create imposters files, using json
* Configure response headers.
* Configure CORS.
* Run the tool using flags or using a config file.
* TODO: ...

## Installing

Since it is built using Golang, so you can easily install it using `go get`

```sh
go install alsritter.icu/middlebaby
```

## Using Middlebaby from the command line

TODO: ...

```
$ middlebaby -h 
a Mock server tool.

Usage:
  middlebaby [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  server      run Mock serve

Flags:
      --app string         Startup app path
      --config string      config file (default is $WORKSPACE/.middlebaby.yaml)
  -h, --help               help for middlebaby
      --log-level string   Log level (default "INFO")

Use "middlebaby [command] --help" for more information about a command.
```


## Using Middlebaby by config file
If we want a more permanent configuration, we could use the option --config to specify the location of a configuration file.

```sh
middlebaby server --log-level="TRACE" --app="./${BIN_FILE}" --config=$WORKSPACE/.middlebaby.yaml
```

The config file must be a YAML file with the following structure.


```yml
# proxy port
port: 7689 
# HTTP requests that need to be mock
httpFiles:
  - ./tests/http.mock.json
# Whether to listen for file changes
watcher: true
# whether the missed mock allows real requests
enableDirect: true
# Set the suffix name of your Task file
taskFileSuffix: "case.json" # e.g., test.case.json
cors: 
  methods: ["GET"]
  headers: ["Content-Type"]
  exposed_headers: ["Cache-Control"]
  origins: ["*"]
  allow_credentials: true
storage: # Configure your mysql and Redis
  mysql:
    port: "3306"
    host: "127.0.0.1"
    database: "test_mb"
    username: "root"
    password: "123456"
    local: "Asia/Shanghai"
    charset: "utf8mb4"
  redis:
    port: "6379"
    host: "127.0.0.1"
    auth: "123456"
    db: 0
caseFiles:
  - "./tests/cases/*.case.json"
```

http mock file

```json
[
  {
    "request": {
      "method": "GET",
      "url": "https://example.org/get",
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
      "body": "{\"name\":\"John\",\"color\":\"Purples\",\"age\":55}"
    }
  },
  {
    "request": {
      "method": "GET",
      "url": "https://example02.org/get"
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
      "body": "{\"data\":{\"name\":\"Alice\",\"color\":\"Blue\",\"age\":18}}"
    }
  }
]
```

http task cases

```json
{
  "serviceType": "http",
  "serviceMethod": "GET",
  "serviceURL": "http://localhost:8011/example",
  "serviceDescription": "This is the first test service",
  "serviceName": "Test access to the Mock's external service",
  "cases": [
    {
      "name": "first case",
      "setup": {
        "mysql": [],
        "redis": [],
        "http": []
      },
      "request": {
        "header": {},
        "data": {}
      },
      "assert": {
        "response": {
          "data": {
            "name": "John",
            "color": "Purples",
            "age": 55
          }
        },
        "mysql": [],
        "redis": []
      },
      "teardown": {
        "mysql": [],
        "redis": []
      }
    }
  ]
}
```

## How to use

TODO: ...