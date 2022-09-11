/*
 Copyright (C) 2022 alsritter

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU Affero General Public License as
 published by the Free Software Foundation, either version 3 of the
 License, or (at your option) any later version.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU Affero General Public License for more details.

 You should have received a copy of the GNU Affero General Public License
 along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package storageprovider

import (
	"errors"
	"time"

	"github.com/spf13/pflag"

	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/go-redis/redis"
	"github.com/go-sql-driver/mysql"
	mysql_driver "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Config struct {
	EnableDocker bool  `json:"enableDocker"`
	Mysql        Mysql `yaml:"mysql"`
	Redis        Redis `yaml:"redis"`
}

func NewConfig() *Config {
	return &Config{
		EnableDocker: false,
		Mysql: Mysql{
			Enabled:  true,
			Port:     "3306",
			Host:     "127.0.0.1",
			Database: "",
			Username: "root",
			Password: "123456",
			Local:    "",
			Charset:  "",
		},
		Redis: Redis{
			Enabled: true,
			Port:    "6379",
			Host:    "127.0.0.1",
			Auth:    "",
			DB:      0,
		},
	}
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {}

func (c *Config) Validate() error {
	if !c.Mysql.Enabled {
		return nil
	}

	cfg := mysql.NewConfig()
	cfg.User = c.Mysql.Username
	cfg.Passwd = c.Mysql.Password
	cfg.Net = "tcp"
	cfg.Addr = c.Mysql.Host + ":" + c.Mysql.Port
	cfg.DBName = c.Mysql.Database
	cfg.Loc, _ = time.LoadLocation(c.Mysql.Local)
	cfg.ParseTime = true
	cfg.Params = map[string]string{"charset": c.Mysql.Charset}

	if _, err := mysql.ParseDSN(cfg.FormatDSN()); err != nil {
		return errors.New("[storage] check your mysql database configuration")
	}

	return nil
}

type Provider interface {
	GetMysqlCon() (*gorm.DB, error)
	GetRedisCon() (*redis.Client, error)
}

type Manager struct {
	cfg *Config
	log logger.Logger
}

func New(log logger.Logger, cfg *Config) Provider {
	return &Manager{
		cfg: cfg,
		log: log.NewLogger("storage"),
	}
}

func (s *Manager) GetMysqlCon() (*gorm.DB, error) {
	if !s.cfg.Mysql.Enabled {
		return nil, nil
	}

	if s.cfg.Mysql.Host == "" {
		return nil, errors.New(" MySQL The configuration information is incomplete. Check whether you do not need to rely on MySQL")
	}
	return gorm.Open(mysql_driver.Open(s.toMysqlConfig().FormatDSN()), &gorm.Config{})
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

func (s *Manager) GetRedisCon() (*redis.Client, error) {
	if !s.cfg.Redis.Enabled {
		return nil, nil
	}

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
