package mysql

import (
	"fmt"

	"github.com/alsritter/middlebaby/pkg/caseprovider"
	"github.com/alsritter/middlebaby/pkg/pluginregistry"
	"github.com/alsritter/middlebaby/pkg/util/assert"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"gorm.io/gorm"
	db_logger "gorm.io/gorm/logger"
)

type mysqlAssertPlugin struct {
	db  *gorm.DB
	log logger.Logger
}

func New(db *gorm.DB, log logger.Logger) pluginregistry.AssertPlugin {
	db.Logger = db.Logger.LogMode(db_logger.Silent)
	if log.GetCurrentLevel() == "trace" {
		db.Logger = db.Logger.LogMode(db_logger.Info)
	}

	return &mysqlAssertPlugin{db: db, log: log.NewLogger("plugin.assert.mysql")}
}

// Name implements pluginregistry.AssertPlugin
func (*mysqlAssertPlugin) Name() string {
	return "mysqlAssertPlugin"
}

// Assert implements pluginregistry.AssertPlugin
func (m *mysqlAssertPlugin) Assert(ca []caseprovider.CommonAssert) error {
	for _, sqlAssert := range ca {
		if result, err := m.run(sqlAssert.Actual); err != nil {
			return err
		} else if len(result) <= 0 {
			return fmt.Errorf("no result is found: %s", sqlAssert.Actual)

			// this result[0] returns a map
		} else if err := assert.So(m.log, "MySQL data assert", result[0], sqlAssert.Expected); err != nil {
			return err
		}
	}
	return nil
}

func (m *mysqlAssertPlugin) run(sql string) (result []map[string]interface{}, err error) {
	err = m.db.Raw(sql).Find(&result).Error
	m.log.Trace(nil, "RUN MySQL: %s %v \n", sql, result)
	return
}
