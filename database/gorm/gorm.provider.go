package gorm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/duolacloud/microbase/database"
	"github.com/duolacloud/microbase/database/gorm/opentracing"
	"github.com/duolacloud/microbase/multitenancy"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/micro/go-micro/v2/config"
)

func NewGormTenancy(config config.Config, entityMap database.EntityMap) (multitenancy.Tenancy, error) {
	driver := config.Get("db", "driver").String("")
	connectionString := config.Get("db", "connection_string").String("")

	if len(driver) == 0 {
		return nil, errors.New("driver is empty")
	}

	if len(connectionString) == 0 {
		return nil, errors.New("connection_string is empty")
	}

	masterDB, err := gorm.Open(driver, connectionString)
	if err != nil {
		return nil, err
	}
	defer masterDB.Close()

	masterDB.LogMode(true)
	masterDB.DB().SetMaxIdleConns(1)
	masterDB.DB().SetConnMaxLifetime(3 * time.Minute)

	dbName := "" // 从连接中获取

	var clientCreateFn = func(ctx context.Context, tenantName string) (multitenancy.Resource, error) {
		dsn := strings.Replace(connectionString, dbName, DBName(dbName, tenantName), 1)

		db, err := gorm.Open(driver, dsn)
		if err != nil {
			return nil, err
		}

		// defer db.Close()
		db.LogMode(true)
		db.DB().SetMaxIdleConns(10)
		db.DB().SetConnMaxLifetime(3 * time.Minute)

		addAutoCallbacks(db)

		opentracing.AddGormCallbacks(db)

		autoMigrate(entityMap, db)
		return db, nil
	}

	tenancy := multitenancy.NewCachedTenancy(clientCreateFn, clientCloseFunc)

	return tenancy, nil
}

var clientCloseFunc = func(resource multitenancy.Resource) {

}

// DBName returns the prefixed database name in order to avoid collision with MySQL internal databases.
func DBName(prefix string, tenantName string) string {
	if len(tenantName) == 0 {
		return prefix
	}
	return fmt.Sprintf("%s_%s", prefix, tenantName)
}

// FromDBName returns the source name of the tenant.
func FromDBName(serviceName string, name string) string {
	return strings.TrimPrefix(name, fmt.Sprintf("%s_", serviceName))
}

func DBFromContext(tenancy multitenancy.Tenancy, ctx context.Context) (*gorm.DB, error) {
	tenantName, _ := multitenancy.FromContext(ctx)

	db, err := tenancy.ResourceFor(ctx, tenantName)
	if err != nil {
		return nil, err
	}
	return db.(*gorm.DB), nil
}

func addAutoCallbacks(db *gorm.DB) {
	// 替换替换默认的钩子
	db.Callback().Create().Replace("gorm:update_time_stamp", updateTimeForCreateCallback)
	db.Callback().Update().Replace("gorm:update_time_stamp", updateTimeForUpdateCallback)
	db.Callback().Delete().Replace("gorm:delete", deleteCallback)
}

func autoMigrate(entityMap database.EntityMap, db *gorm.DB) error {
	// ctx, span := trace.StartSpan(ctx, "tenancy.Migrate")
	// defer span.End()
	entities := entityMap.GetEntities()
	if db = db.AutoMigrate(entities...); db.Error != nil {
		// TODO m.logger.For(ctx).Error("tenancy migrate", zap.Error(err))
		// span.SetStatus(trace.Status{Code: trace.StatusCodeUnknown, Message: err.Error()})
		return fmt.Errorf("running tenancy migration: %w", db.Error)
	}
	return nil
}
