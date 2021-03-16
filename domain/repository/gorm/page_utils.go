package gorm

import (
	_gorm "github.com/jinzhu/gorm"
)

func pageQuery(queryHandler *_gorm.DB, pageNo int, pageSize int, resultPtr interface{}) (total int64, err error) {
	limit, offset := getLimitOffset(pageNo-1, pageSize)

	queryHandler.Count(&total)
	queryHandler.Limit(limit).Offset(offset).Find(resultPtr)
	if err = queryHandler.Error; err != nil {
		return
	}

	return
}

func getLimitOffset(pageNo, pageSize int) (limit, offset int) {
	if pageNo < 0 {
		pageNo = 0
	}

	if pageSize < 1 {
		pageSize = 20
	}
	return pageSize, pageNo * pageSize
}
