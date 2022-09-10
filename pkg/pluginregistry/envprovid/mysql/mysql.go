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

package envmysql

import (
	"github.com/alsritter/middlebaby/pkg/pluginregistry"
	"github.com/alsritter/middlebaby/pkg/storageprovider"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/hashicorp/go-multierror"
	"gorm.io/gorm"
	db_logger "gorm.io/gorm/logger"
)

type MySQLEnvPlugin struct {
	db  *gorm.DB
	log logger.Logger
}

func New(storage storageprovider.Provider, log logger.Logger) pluginregistry.EnvPlugin {
	db, err := storage.GetMysqlCon()
	if err != nil {
		log.Error(nil, "MySQLEnvPlugin init failed: %v", err)
	}

	db.Logger = db.Logger.LogMode(db_logger.Silent)
	if log.GetCurrentLevel() == "trace" {
		db.Logger = db.Logger.LogMode(db_logger.Info)
	}

	return &MySQLEnvPlugin{db: db, log: log.NewLogger("plugin.env.mysql")}
}

func (*MySQLEnvPlugin) GetTypeName() string {
	return "mysql"
}

func (*MySQLEnvPlugin) Name() string {
	return "MySQLEnvPlugin"
}

func (m *MySQLEnvPlugin) Run(commands []string) error {
	var errs error
	for _, cmd := range commands {
		_, err := m.run(cmd)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}

func (m *MySQLEnvPlugin) run(sql string) (result []map[string]interface{}, err error) {
	err = m.db.Raw(sql).Find(&result).Error
	m.log.Trace(nil, "RUN MySQL: %s %v \n", sql, result)
	return
}
