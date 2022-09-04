package envredis

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/alsritter/middlebaby/pkg/storageprovider"

	"github.com/alsritter/middlebaby/pkg/pluginregistry"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/go-redis/redis"
	"github.com/hashicorp/go-multierror"
)

type RedisEnvPlugin struct {
	log logger.Logger
	rc  *redis.Client
}

func New(storage storageprovider.Provider, log logger.Logger) pluginregistry.EnvPlugin {
	rc, err := storage.GetRedisCon()
	if err != nil {
		log.Error(nil, "redisAssertPlugin init failed: %v", err)
	}
	return &RedisEnvPlugin{rc: rc, log: log.NewLogger("plugin.env.redis")}
}

func (*RedisEnvPlugin) Name() string {
	return "RedisEnvPlugin"
}

func (*RedisEnvPlugin) GetTypeName() string {
	return "redis"
}

func (r *RedisEnvPlugin) Run(commands []string) error {
	var errs error
	for _, cmd := range commands {
		_, err := r.run(cmd)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}

func (r *RedisEnvPlugin) run(cmd string) (result interface{}, err error) {
	// formatting command.
	cmdList := r.redisParse(cmd)
	r.log.Trace(nil, "redis parse list: %v", cmdList)
	if len(cmdList) <= 0 {
		return nil, nil
	}

	commandName := strings.ToLower(cmdList[0])
	switch commandName {
	case "get":
		result, err = r.rc.Get(cmdList[1]).Result()
	case "hgetall":
		result, err = r.rc.HGetAll(cmdList[1]).Result()
	case "set":
		if len(cmdList) < 3 {
			r.log.Error(nil, "redis command error:", cmd)
			return nil, fmt.Errorf("the redis command format is incorrect")
		}
		result, err = r.rc.Set(cmdList[1], strings.Join(cmdList[2:], ""), -1).Result()
	default:
		var icmds []interface{}
		for _, v := range cmdList {
			icmds = append(icmds, v)
		}
		// TODO: the result returned by the do method is interface{}([]byte), inconsistent with the format expected by the default test case. so here need reflection conversion type.
		result, err = r.rc.Do(icmds...).Result()
	}
	if err == redis.Nil {
		result, err = redis.Nil.Error(), nil
	}

	r.log.Trace(nil, "RUN Redis: %s result:%v %v", cmd, result, err)
	return
}

// formatting command.
// separated by Spaces, if there are double quotation marks, they are not separated and then finally get rid of the "
// example: SET testkey "this is value."  => ([0] = SET, [1] = testkey, [2] = this is value.)
func (r *RedisEnvPlugin) redisParse(cmd string) []string {
	cmd = strings.TrimSpace(cmd)
	exp, _ := regexp.Compile("\"[^\"]+\"|\\S+")
	cmdList := exp.FindAllString(cmd, -1)
	for i := 0; i < len(cmdList); i++ {
		// get rid of the "
		cmdList[i] = strings.Trim(cmdList[i], "\"")
	}
	return cmdList
}
