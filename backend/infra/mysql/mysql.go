package mysql

import (
	"q-dev/conf"
	"q-dev/infra/mysql/dao"
	"q-dev/infra/mysql/model"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(cfg conf.MySQLConfig) error {
	var err error
	DB, err = gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	if err := DB.Use(otelgorm.NewPlugin()); err != nil {
		return err
	}

	dao.SetDefault(DB)
	return nil
}

func Migrate() error {
	if DB == nil {
		return gorm.ErrInvalidDB
	}
	return DB.AutoMigrate(
		&model.HelloWorld{},
	)
}

func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
