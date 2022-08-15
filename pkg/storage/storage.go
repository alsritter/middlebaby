package storage

import (
	"errors"
	"time"

	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/go-redis/redis"
	"github.com/go-sql-driver/mysql"
	mysql_driver "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Config struct {
	Mysql Mysql `yaml:"mysql"`
	Redis Redis `yaml:"redis"`
}

type Mysql struct {
	Port     string `yaml:"port"`
	Host     string `yaml:"host"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Local    string `yaml:"local"`
	Charset  string `yaml:"charset"`
}

type Redis struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
	Auth string `yaml:"auth"`
	DB   int    `yaml:"db"`
}

type Provider interface {
	GetMysqlCon() (*gorm.DB, error)
	GetRedisCon() (*redis.Client, error)
}

type Manager struct {
	cfg    *Config
	logger logger.Logger
}

func New(cfg *Config, logger logger.Logger) (Provider, error) {
	return &Manager{
		cfg:    cfg,
		logger: logger,
	}, nil
}

func (s *Manager) toMysqlConfig() *mysql.Config {
	cfg := mysql.NewConfig()
	cfg.User = s.cfg.Mysql.Username
	cfg.Passwd = s.cfg.Mysql.Password
	cfg.Net = "tcp"
	cfg.Addr = s.cfg.Mysql.Host + ":" + s.cfg.Mysql.Port
	cfg.DBName = s.cfg.Mysql.Database
	cfg.Loc, _ = time.LoadLocation(s.cfg.Mysql.Local)
	cfg.ParseTime = true
	cfg.Params = map[string]string{"charset": s.cfg.Mysql.Charset}
	return cfg
}

func (s *Manager) GetMysqlCon() (*gorm.DB, error) {
	if s.cfg.Mysql.Host == "" {
		return nil, errors.New(" MySQL The configuration information is incomplete. Check whether you do not need to rely on MySQL")
	}
	return gorm.Open(mysql_driver.Open(s.toMysqlConfig().FormatDSN()), &gorm.Config{})
}

func (s *Manager) GetRedisCon() (*redis.Client, error) {
	if s.cfg.Redis.Host == "" {
		return nil, errors.New(" Redis The configuration information is incomplete. Check whether Redis is not required")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     s.cfg.Redis.Host + ":" + s.cfg.Redis.Port,
		Password: s.cfg.Redis.Auth,
		DB:       s.cfg.Redis.DB,
	})
	return rdb, rdb.Ping().Err()
}
