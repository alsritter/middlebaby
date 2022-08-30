package mysql

import (
	"fmt"
	"github.com/alsritter/middlebaby/pkg/caseprovider"
	"github.com/alsritter/middlebaby/pkg/pluginregistry"
	"github.com/alsritter/middlebaby/pkg/util/assert"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"gorm.io/gorm"
	dblogger "gorm.io/gorm/logger"
)

type mysqlAssertPlugin struct {
	db  *gorm.DB
	log logger.Logger
}

func New(db *gorm.DB, log logger.Logger) pluginregistry.AssertPlugin {
	db.Logger = db.Logger.LogMode(dblogger.Silent)
	if log.GetCurrentLevel() == "trace" {
		db.Logger = db.Logger.LogMode(dblogger.Info)
	}

	return &mysqlAssertPlugin{db: db, log: log.NewLogger("plugin.assert.mysql")}
}

func (m *mysqlAssertPlugin) Name() string {
	return "mysqlAssertPlugin"
}

func (m *mysqlAssertPlugin) GetTypeName() string {
	return "mysql"
}

// Assert run mysql assertprovid.
func (m *mysqlAssertPlugin) Assert(asserts []caseprovider.CommonAssert) error {
	for _, commonAssert := range asserts {
		if result, err := m.run(commonAssert.Actual); err != nil {
			return err
		} else if len(result) <= 0 {
			return fmt.Errorf("no result is found: %s", commonAssert.Actual)

			// this result[0] returns a map
		} else if err := assert.So(m.log, "MySQL data assert", result[0], commonAssert.Expected); err != nil {
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
