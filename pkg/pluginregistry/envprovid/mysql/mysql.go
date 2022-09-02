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
		log.Error(nil, "MySQLEnvPlugin init failed: %w", err)
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
