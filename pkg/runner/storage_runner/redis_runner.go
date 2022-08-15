package storage_runner

import (
	"fmt"
	"github.com/alsritter/middlebaby/pkg/runner"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"regexp"
	"strings"

	"github.com/go-redis/redis"
)

var _ runner.RedisRunner = (*redisInstance)(nil)
var _ runner.RedisRunner = (*defaultRedisInstance)(nil)

type redisInstance struct {
	log         logger.Logger
	redisClient *redis.Client
}

// NewRedisRunner return a redis runner.
func NewRedisRunner(redisClient *redis.Client) runner.RedisRunner {
	if redisClient == nil {
		return &defaultRedisInstance{}
	}
	return &redisInstance{redisClient: redisClient}
}

// formatting command.
// separated by Spaces, if there are double quotation marks, they are not separated and then finally get rid of the "
// example: SET testkey "this is value."  => ([0] = SET, [1] = testkey, [2] = this is value.)
func (r *redisInstance) redisParse(cmd string) []string {
	cmd = strings.TrimSpace(cmd)
	exp, _ := regexp.Compile("\"[^\"]+\"|\\S+")
	cmdList := exp.FindAllString(cmd, -1)
	for i := 0; i < len(cmdList); i++ {
		// get rid of the "
		cmdList[i] = strings.Trim(cmdList[i], "\"")
	}
	return cmdList
}

func (r *redisInstance) Run(cmd string) (result interface{}, err error) {
	// formatting command.
	cmdList := r.redisParse(cmd)
	r.log.Trace(nil, "redis parse list: %v", cmdList)
	if len(cmdList) <= 0 {
		return nil, nil
	}

	commandName := strings.ToLower(cmdList[0])
	switch commandName {
	case "get":
		result, err = r.redisClient.Get(cmdList[1]).Result()
	case "hgetall":
		result, err = r.redisClient.HGetAll(cmdList[1]).Result()
	case "set":
		if len(cmdList) < 3 {
			r.log.Error(nil, "redis command error:", cmd)
			return nil, fmt.Errorf("the redis command format is incorrect")
		}
		result, err = r.redisClient.Set(cmdList[1], strings.Join(cmdList[2:], ""), -1).Result()
	default:
		var icmds []interface{}
		for _, v := range cmdList {
			icmds = append(icmds, v)
		}
		// TODO: the result returned by the do method is interface{}([]byte), inconsistent with the format expected by the default test case. so here need reflection conversion type.
		result, err = r.redisClient.Do(icmds...).Result()
	}
	if err == redis.Nil {
		result, err = redis.Nil.Error(), nil
	}

	r.log.Trace(nil, "RUN Redis: %s result:%v %v", cmd, result, err)
	return
}

type defaultRedisInstance struct {
	log logger.Logger
}

func (d *defaultRedisInstance) Run(cmd string) (result interface{}, err error) {
	err = fmt.Errorf("information is not configured in the configuration file, Confirm whether the Redis statement needs to be executed ")
	return
}
