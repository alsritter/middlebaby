package task

import (
	"alsritter.icu/middlebaby/internal/log"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var _ (MysqlRunner) = (*mysqlInstance)(nil)
var _ (MysqlRunner) = (*defaultMysqlInstance)(nil)

// return a mysql runner.
func NewMysqlRunner(db *gorm.DB) MysqlRunner {
	if db == nil {
		return &defaultMysqlInstance{}
	}

	db.Logger = db.Logger.LogMode(logger.Silent)
	if log.GetCurrentLevel() == log.TraceLevel {
		db.Logger = db.Logger.LogMode(logger.Info)
	}

	return &mysqlInstance{db: db}
}

type mysqlInstance struct {
	db *gorm.DB
}

func (m *mysqlInstance) Run(sql string) (result []map[string]interface{}, err error) {
	err = m.db.Raw(sql).Find(&result).Error
	log.Tracef("RUN MySQL: %s %v \n", sql, result)
	return
}

// MySQL information is not configured in the configuration file, default instances are generated, and no SQL operations are performed
type defaultMysqlInstance struct {
}

func (d *defaultMysqlInstance) Run(sql string) (result []map[string]interface{}, err error) {
	log.Warn("information is not configured in the configuration file, Confirm whether the SQL statement needs to be executed ?")
	return
}
