
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




