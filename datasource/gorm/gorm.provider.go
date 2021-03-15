package gorm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/duolacloud/microbase/datasource"
	"github.com/duolacloud/microbase/datasource/gorm/opentracing"
	"github.com/duolacloud/microbase/multitenancy"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/micro/go-micro/v2/config"
)

func NewGormTenancy(config config.Config, entityMap database.EntityMap) (multitenancy.Tenancy, error) {
	driver := config.Get("db", "driver").String("")
	connectionString := config.Get("db", "connection_string").String("")

	isolation := config.Get("multitenancy", "isolation").String("schema")

	if len(driver) == 0 {
		return nil, errors.New("driver is empty")
	}

	if len(connectionString) == 0 {
		return nil, errors.New("connection_string is empty")
	}

	defaultDB, err := gorm.Open(driver, connectionString)
	if err != nil {
		return nil, err
	}
	// defer defaultDB.Close()

	defaultDB.LogMode(true)
	defaultDB.DB().SetMaxIdleConns(1)
	defaultDB.DB().SetConnMaxLifetime(3 * time.Minute)

	addAutoCallbacks(defaultDB)
	opentracing.AddGormCallbacks(defaultDB)

	var clientCreateFn func(ctx context.Context, tenantId string) (multitenancy.Resource, error)
	var clientCloseFunc func(resource multitenancy.Resource)
	switch isolation {
	case "schema":
		clientCreateFn = func(ctx context.Context, tenantId string) (multitenancy.Resource, error) {
			db := defaultDB // gorm.Open(driver, connectionString)

			autoMigrate(tenantId, entityMap, db)
			return db, nil
		}

		clientCloseFunc = func(resource multitenancy.Resource) {}
	case "database":
		dbName := "" // 从连接中获取

		clientCreateFn = func(ctx context.Context, tenantId string) (multitenancy.Resource, error) {
			dsn := strings.Replace(connectionString, dbName, DBName(dbName, tenantId), 1)

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

			autoMigrate(tenantId, entityMap, db)
			return db, nil
		}

		clientCloseFunc = func(resource multitenancy.Resource) {}
	}

	tenancy := multitenancy.NewCachedTenancy(clientCreateFn, clientCloseFunc)

	return tenancy, nil
}

// DBName returns the prefixed database name in order to avoid collision with MySQL internal databases.
func DBName(prefix string, tenantId string) string {
	if len(tenantId) == 0 {
		return prefix
	}
	return fmt.Sprintf("%s_%s", prefix, tenantId)
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

func TableName(tableName string, tenantId string) string {
	if len(tenantId) == 0 {
		return tableName
	}

	return fmt.Sprintf("%s_%s", tableName, tenantId)
}

func autoMigrate(tenantId string, entityMap database.EntityMap, db *gorm.DB) error {
	// ctx, span := trace.StartSpan(ctx, "tenancy.Migrate")
	// defer span.End()
	entities := entityMap.GetEntities()

	db = db.Unscoped()
	for _, entity := range entities {
		scope := db.NewScope(entity)
		tableName := TableName(scope.TableName(), tenantId)
		db = db.Table(tableName).AutoMigrate(entity)
	}

	return nil
}
