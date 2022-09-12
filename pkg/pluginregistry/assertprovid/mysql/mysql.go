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

package mysql

import (
	"fmt"

	"github.com/alsritter/middlebaby/pkg/pluginregistry"
	"github.com/alsritter/middlebaby/pkg/storageprovider"
	"github.com/alsritter/middlebaby/pkg/types/mbcase"
	"github.com/alsritter/middlebaby/pkg/util/assert"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"gorm.io/gorm"
	dblogger "gorm.io/gorm/logger"
)

type mysqlAssertPlugin struct {
	db  *gorm.DB
	log logger.Logger
}

func New(storage storageprovider.Provider, log logger.Logger) pluginregistry.AssertPlugin {
	db, err := storage.GetMysqlCon()
	if err != nil {
		log.Error(nil, "mysqlAssertPlugin init failed: %v", err)
	}

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
func (m *mysqlAssertPlugin) Assert(_ *mbcase.Response, asserts []mbcase.CommonAssert) error {
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
