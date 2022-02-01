The structure of the Task module is as follows:

```
          TaskService
               |
               |
            Runner (Pass Runner to the following function to execute)
               |
               v
         HttpTaskRunner (Find CaseName by InterfaceName)
               |
               |
   +-----------+------FindCase 
   |           |           |
   |     InterfaceName     |
   |           |           |
   |     +-----v----+      |
   |     |          |      |
   |     |          |      |
   |  CaseName    CaseName |
   |     |          |      |
   +-----+----------+------+
         |          |
         |          |
         v          v
httpTaskCase      httpTaskCase (Execute a Case)
```

1. `Runner`(runner.go): It is the interface that actually performs the requests, including mysql, Redis, HTTP, GRPC requests.

2. `httpTaskCase` or `grpcTaskCase`(grpc_task.go, http_task.go): They are used to execute a specific Case under a Task. (In fact, it is also executed by calling Runner)

3. `HttpTaskRunner` or `GRpcTaskRunner`(grpc_runner.go, http_runner.go): They are save all http task or grpc task. It is used to find the corresponding CaseName, which is handed to `httpTaskCase` or `grpcTaskCase` for execution.

4. `TaskService` : Interface test entry service, which is used to organize the services above