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

Since it is built using Golang, so you can easily install it using `go install`

```sh
go install github.com/alsritter/middlebaby@latest
```

## Using Middlebaby from the command line

TODO: ...

```
$ middlebaby -h 
a auto mock tool.

Usage:
  middlebaby [flags]
  middlebaby [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  init        init config file
  serve       start the mock server

Flags:
  -h, --help      help for middlebaby
  -v, --version   version for middlebaby

Use "middlebaby [command] --help" for more information about a command.
```


## Using Middlebaby by config file
use Makfile.

```sh
make run-http
```

The config file must be a YAML file with the following structure.


```yml
log:
  prefix: true
  level: debug
target:
  appPath: "./target"
mock:
  enableDirect: true
  mockPort: 9090
task:
  targetServeAdder: "127.0.0.1:8011"
  closeTearDown: false
storage:
  enabledocker: false
  mysql:
    enabled: true
    port: "3306"
    host: "127.0.0.1"
    database: "test_mb"
    username: "root"
    password: "123456"
    local: "Asia/Shanghai"
    charset: "utf8mb4"
  redis:
    enabled: true
    port: "6379"
    host: "127.0.0.1"
    auth: "123456"
    db: 0
case:
  taskFileSuffix: .case.json
  caseFiles: 
      - "./tests/cases/**/*.case.json"
  watcherCases: true
  mockFiles:
      - ./tests/http.mock.json
  watcherMock: true
proto:
  protoimportpaths: []
  sync:
    enable: false
    storagedir: ""
    repository: []
web:
  port: 6060
```

http mock file

```json
[
  {
    "request": {
      "protocol": "http",
      "method": "GET",
      "host": "example.org",
      "path": "/get",
      "params": {
        "name": "John",
        "age": "55"
      }
    },
    "response": {
      "status": 200,
      "header": {
        "Content-Type": ["application/json"]
      },
      "body": "{\"name\":\"John\",\"color\":\"Purples\",\"age\":55}",
      "trailer": {},
      "delay": {
        "delay": 100,
        "offset": 50
      }
    }
  },
  {
    "request": {
      "protocol": "http",
      "method": "GET",
      "host": "https://example02.org",
      "path": "/get"
    },
    "response": {
      "status": 200,
      "header": {
        "Content-Type": ["application/json"]
      },
      "body": "{\"name\":\"Alice\",\"color\":\"Blue\",\"age\":18}",
      "trailer": {},
      "delay": {
        "delay": 300,
        "offset": 200
      }
    }
  }
]
```

http task cases

```json
{
  "protocol": "http",
  "serviceMethod": "GET",
  "serviceName": "Test access to the Mock's external service",
  "serviceDescription": "This is the first test service",
  "servicePath": "http://localhost:8011/example",
  "serviceProtoFile": "",
  "setup": [
    {
      "typeName": "",
      "commands": []
    }
  ],
  "mocks": [
    {
      "request": {
        "protocol": "http",
        "method": "GET",
        "host": "example.org",
        "path": "/get",
        "query": {
          "name": ["John"],
          "age": ["55"]
        }
      },
      "response": {
        "status": 200,
        "header": {
          "Content-Type": ["application/json"]
        },
        "body": "{\"name\":\"John\",\"color\":\"Purples\",\"age\":88}",
        "trailer": {},
        "delay": {
          "delay": 300,
          "offset": 50
        }
      }
    }
  ],
  "teardown": [
    {
      "typeName": "",
      "commands": []
    }
  ],
  "cases": [
    {
      "name": "first case-fail",
      "description": "Test cases that will fail~",
      "setup": [],
      "mocks": [],
      "request": {
        "header": {},
        "query": {},
        "data": null
      },
      "assert": {
        "response": {
          "data": {
            "name": "John",
            "color": "Purples",
            "age": 55
          }
        },
        "otherAsserts": []
      },
      "teardown": []
    },
    {
      "name": "first case-success",
      "description": "",
      "setup": [],
      "mocks": [
        {
          "request": {
            "protocol": "http",
            "method": "GET",
            "host": "example.org",
            "path": "/get",
            "query": {
              "name": ["John"],
              "age": ["55"]
            }
          },
          "response": {
            "status": 200,
            "header": {
              "Content-Type": ["application/json"]
            },
            "body": "{\"name\":\"John\",\"color\":\"Purples\",\"age\":55}",
            "trailer": {},
            "delay": {
              "delay": 300,
              "offset": 50
            }
          }
        }
      ],
      "request": {
        "header": {},
        "query": {},
        "data": null
      },
      "assert": {
        "response": {
          "data": {
            "name": "John",
            "color": "Purples",
            "age": 55
          }
        },
        "otherAsserts": []
      },
      "teardown": []
    },
    {
      "name": "first case-regex-success",
      "description": "",
      "setup": [],
      "mocks": [
        {
          "request": {
            "protocol": "http",
            "method": "GET",
            "host": "example.org",
            "path": "/get",
            "query": {
              "name": ["John"],
              "age": ["55"]
            }
          },
          "response": {
            "status": 200,
            "header": {
              "Content-Type": ["application/json"]
            },
            "body": "{\"name\":\"John\",\"color\":\"Purples\",\"age\":55}",
            "trailer": {},
            "delay": {
              "delay": 300,
              "offset": 50
            }
          }
        }
      ],
      "request": {
        "header": {},
        "query": {},
        "data": null
      },
      "assert": {
        "response": {
          "data": {
            "name": "@regExp:^J.*",
            "color": "Purples",
            "age": 55
          }
        },
        "otherAsserts": []
      },
      "teardown": []
    },
    {
      "name": "first case-js-success",
      "description": "",
      "setup": [],
      "mocks": [
        {
          "request": {
            "protocol": "http",
            "method": "GET",
            "host": "example.org",
            "path": "/get",
            "query": {
              "name": ["John"],
              "age": ["55"]
            }
          },
          "response": {
            "status": 200,
            "header": {
              "Content-Type": ["application/json"]
            },
            "body": "{\"name\":\"John\",\"color\":\"Purples\",\"age\":55}",
            "trailer": {},
            "delay": {
              "delay": 300,
              "offset": 50
            }
          }
        }
      ],
      "request": {
        "header": {},
        "query": {},
        "data": null
      },
      "assert": {
        "response": {
          "data": {
            "name": "@regExp:^J.*",
            "color": "Purples",
            "age": 55
          }
        },
        "otherAsserts": [
          {
            "typeName": "js",
            "actual": "assert.data.statusCode == 200",
            "expected": true
          },
          {
            "typeName": "js",
            "actual": "assert.data.data.color == 'Purples'",
            "expected": true
          }
        ]
      },
      "teardown": []
    }
  ]
}
```

## How to use

TODO: ...