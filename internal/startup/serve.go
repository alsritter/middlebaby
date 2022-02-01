package startup

import (
	"errors"
	"time"

	"alsritter.icu/middlebaby/internal/event"
	"alsritter.icu/middlebaby/internal/log"
	"alsritter.icu/middlebaby/internal/proxy"
	"alsritter.icu/middlebaby/internal/startup/plugin"
	"alsritter.icu/middlebaby/internal/task"
	"github.com/go-redis/redis"
	"github.com/go-sql-driver/mysql"
	mysql_driver "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Serve struct {
	taskService *task.TaskService
}

func NewCaseServe(env plugin.Env, mockCenter proxy.MockCenter) *Serve {
	taskService, err := task.NewTaskService(env, mockCenter, newRunner(env))

	if err != nil {
		log.Fatal(err)
		return nil
	}

	return &Serve{
		taskService: taskService,
	}
}

func (s *Serve) Run() {
	taskCaseMap := s.taskService.GetAllTestCase()
	mustRunTearDown := true
	for _, testCaseType := range []string{task.TestCaseTypeGRpc, task.TestCaseTypeHTTP} {
		t := taskCaseMap[testCaseType]
		if t == nil {
			continue
		}

		interfaceList := t.GetTaskCaseTree()
		for _, iFace := range interfaceList {
			for _, caseName := range iFace.CaseList {
				if err := s.taskService.Run(testCaseType, caseName, &mustRunTearDown); err != nil {
					log.Error("execute failure ", caseName, err.Error())
				} else {
					log.Debug("execute successfully ", caseName)
				}
			}
		}
	}

	// shutdown after execution
	event.Bus.Publish(event.CLOSE)
}

func newRunner(env plugin.Env) task.Runner {
	db, err := getMysqlCon(env)
	if err != nil {
		log.Error("Failed to connect to the MySQL database:", err.Error())
	}

	redisPool, err := getRedisCon(env)
	if err != nil {
		log.Error("Failed to connect to the Redis:", err.Error())
	}
	runner, err := task.NewRunner(task.NewMysqlRunner(db), task.NewRedisRunner(redisPool))
	if err != nil {
		log.Fatal("Failed to initialize the running environment:", err)
	}
	return runner
}

func toMysqlConfig(env plugin.Env) *mysql.Config {
	cfg := mysql.NewConfig()
	cfg.User = env.GetConfig().Storage.Mysql.Username
	cfg.Passwd = env.GetConfig().Storage.Mysql.Password
	cfg.Net = "tcp"
	cfg.Addr = env.GetConfig().Storage.Mysql.Host + ":" + env.GetConfig().Storage.Mysql.Port
	cfg.DBName = env.GetConfig().Storage.Mysql.Database
	cfg.Loc, _ = time.LoadLocation(env.GetConfig().Storage.Mysql.Local)
	cfg.ParseTime = true
	cfg.Params = map[string]string{"charset": env.GetConfig().Storage.Mysql.Charset}
	return cfg
}

func getMysqlCon(env plugin.Env) (*gorm.DB, error) {
	if env.GetConfig().Storage.Mysql.Host == "" {
		return nil, errors.New(" MySQL The configuration information is incomplete. Check whether you do not need to rely on MySQL")
	}
	return gorm.Open(mysql_driver.Open(toMysqlConfig(env).FormatDSN()), &gorm.Config{})
}

func getRedisCon(env plugin.Env) (*redis.Client, error) {
	if env.GetConfig().Storage.Redis.Host == "" {
		return nil, errors.New(" Redis The configuration information is incomplete. Check whether Redis is not required")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     env.GetConfig().Storage.Redis.Host + ":" + env.GetConfig().Storage.Redis.Port,
		Password: env.GetConfig().Storage.Redis.Auth,
		DB:       env.GetConfig().Storage.Redis.DB,
	})
	return rdb, rdb.Ping().Err()
}
