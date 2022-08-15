package storage_runner

import (
	"fmt"

	"github.com/alsritter/middlebaby/pkg/taskserver"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"gorm.io/gorm"
	db_logger "gorm.io/gorm/logger"
)

var _ taskserver.MysqlRunner = (*mysqlInstance)(nil)
var _ taskserver.MysqlRunner = (*defaultMysqlInstance)(nil)

// NewMysqlRunner return a mysql runner.
func NewMysqlRunner(db *gorm.DB, log logger.Logger) taskserver.MysqlRunner {
	if db == nil {
		return &defaultMysqlInstance{}
	}

	db.Logger = db.Logger.LogMode(db_logger.Silent)
	if log.GetCurrentLevel() == "trace" {
		db.Logger = db.Logger.LogMode(db_logger.Info)
	}

	return &mysqlInstance{db: db, log: log}
}

type mysqlInstance struct {
	db  *gorm.DB
	log logger.Logger
}

func (m *mysqlInstance) Run(sql string) (result []map[string]interface{}, err error) {
	err = m.db.Raw(sql).Find(&result).Error
	m.log.Trace(nil, "RUN MySQL: %s %v \n", sql, result)
	return
}

// information is not configured in the configuration file, default instances are generated, and no SQL operations are performed
type defaultMysqlInstance struct {
}

func (d *defaultMysqlInstance) Run(sql string) (result []map[string]interface{}, err error) {
	err = fmt.Errorf("information is not configured in the configuration file, Confirm whether the SQL statement needs to be executed ")
	return
}
