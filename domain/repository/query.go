package repository

import "errors"

// ErrFilter
var (
	ErrFilter          = errors.New("过滤参数错误")
	ErrFilterValueType = errors.New("过滤值类型错误")
	ErrFilterValueSize = errors.New("过滤值大小错误")
	ErrFilterOperate   = errors.New("过滤操作错误")
)
