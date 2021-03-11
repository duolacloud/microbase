package gorm

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

// 注册删除钩子在删除之前
func deleteCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		var extraOption string
		if str, ok := scope.Get("gorm:delete_option"); ok {
			extraOption = fmt.Sprint(str)
		}

		deleteTimeField, hasDeleteTimeField := scope.FieldByName("dtime")
		deletedField, hasDeletedField := scope.FieldByName("deleted")

		if !scope.Search.Unscoped && hasDeleteTimeField && hasDeletedField {
			scope.Raw(fmt.Sprintf(
				"UPDATE %v SET %v=%v,%v=%v%v%v",
				scope.QuotedTableName(),
				scope.Quote(deleteTimeField.DBName),
				scope.AddToVars(time.Now()),
				scope.Quote(deletedField.DBName),
				scope.AddToVars(1),
				// 组合SQL.比如 update users set a = 1 and b = 2 (where id = x) 后面的sql
				addExtraSpaceIfExist(scope.CombinedConditionSql()),
				// db.Set("gorm:delete_option", "OPTION (OPTIMIZE FOR UNKNOWN)").Delete(&email)
				// DELETE from emails where id=10 OPTION (OPTIMIZE FOR UNKNOWN);
				addExtraSpaceIfExist(extraOption),
			)).Exec()
		} else if !scope.Search.Unscoped && hasDeleteTimeField {
			scope.Raw(fmt.Sprintf(
				"UPDATE %v SET %v=%v%v%v",
				scope.QuotedTableName(),
				scope.Quote(deleteTimeField.DBName),
				scope.AddToVars(time.Now()),
				addExtraSpaceIfExist(scope.CombinedConditionSql()),
				addExtraSpaceIfExist(extraOption),
			)).Exec()
		} else {
			scope.Raw(fmt.Sprintf(
				"DELETE FROM %v%v%v",
				scope.QuotedTableName(),
				addExtraSpaceIfExist(scope.CombinedConditionSql()),
				addExtraSpaceIfExist(extraOption),
			)).Exec()
		}
	}
}

func addExtraSpaceIfExist(str string) string {
	if str != "" {
		return " " + str
	}
	return ""
}

func updateTimeForUpdateCallback(scope *gorm.Scope) {
	// scope.Get(...) 根据入参获取设置了字面值的参数，例如本文中是 gorm:update_column ，它会去查找含这个字面值的字段属性
	if _, ok := scope.Get("gorm:utime"); !ok {
		// scope.SetColumn(...) 假设没有指定 update_column 的字段，我们默认在更新回调设置 ModifiedOn 的值
		scope.SetColumn("utime", time.Now())
	}
}

func updateTimeForCreateCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		nowTime := time.Now()

		// 通过 scope.Fields() 获取所有字段，判断当前是否包含所需字段

		if createTimeField, ok := scope.FieldByName("ctime"); ok {
			if createTimeField.IsBlank { // 可判断该字段的值是否为空
				createTimeField.Set(nowTime)
			}
		}

		if updateTimeField, ok := scope.FieldByName("utime"); ok {
			if updateTimeField.IsBlank { // 可判断该字段的值是否为空
				updateTimeField.Set(nowTime)
			}
		}
	}
}
