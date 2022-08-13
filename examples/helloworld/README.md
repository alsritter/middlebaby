
If you use Docker, you can configure mysql and Redis in the following way.

```sh
docker-compose up -d
```

Otherwise, you need to configure MySQL and Redis as shown in `.middlebaby.yaml`.

```yml
storage:
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
```

Enter the following command in your terminal:

```sh
$ go get github.com/alsritter/middlebaby
$ make buildandrun
```

output:

```
[debug] 2022/02/02 01:09:25 Configuration file to use: /home/alsritter/project/testmb/tests/.middlebaby.yaml
[trace] 2022/02/02 01:09:25 task_service.go:196: task_file.HttpTask{HttpTaskInfo:(*task_file.HttpTaskInfo)(0xc0002123c0), Cases:[]*task_file.HttpTaskCase{(*task_file.HttpTaskCase)(0xc0002e3560)}, InterfaceOperator:(*task_file.InterfaceOperator)(nil)} 
[trace] 2022/02/02 01:09:25 task_service.go:185: loading all task file.
[debug] 2022/02/02 01:09:25 print all http router:
[debug] 2022/02/02 01:09:25 
                        --------------------
                        Method: [GET], err1: <nil>
                        path: /get, err2: <nil>
                        Host: example.org, err3: <nil>
                        queries: [name=John age=55], err4: <nil>
                        --------------------

[debug] 2022/02/02 01:09:25 
                        --------------------
                        Method: [GET], err1: <nil>
                        path: /get, err2: <nil>
                        Host: example02.org, err3: <nil>
                        queries: [], err4: <nil>
                        --------------------

--------------------------------
[trace] 2022/02/02 01:09:27 http_runner.go:58: &{访问测试  {[] [] [] []} {map[] map[] map[]} {{map[] map[age:55 color:Purples name:John] 0} [] []} {[] []}} &{测试访问被 Mock 外部服务 GET 这个是第一个测试服务 http://localhost:8011/example} <nil>
[warn] 2022/02/02 01:09:27 no interfaceOperator found
[debug] 2022/02/02 01:09:27 GET /example HTTP/1.1
Host: localhost:8011

{}
[trace] 2022/02/02 01:09:27 proxy.go:40: target request: example.org GET /get
[trace] 2022/02/02 01:09:27 handler.go:16: hit mock: http://example.org/get?name=John&age=55
[trace] 2022/02/02 01:09:27 handler.go:18: proxy request: GET http://example.org/get?name=John&age=55 HTTP/1.1
Accept-Encoding: gzip
User-Agent: Go-http-client/1.1

{"name":"John","color":"Purples","age":55}
[debug] 2022/02/02 01:09:27 response message: map[Content-Length:[42] Content-Type:[text/plain; charset=utf-8] Date:[Tue, 01 Feb 2022 17:09:27 GMT]] {"name":"John","color":"Purples","age":55} 200 map[age:55 color:Purples name:John] 
[debug] 2022/02/02 01:09:27 execute successfully  访问测试
```