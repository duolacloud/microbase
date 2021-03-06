package logger

import (
	l "github.com/micro/go-micro/v2/logger"
)

func Infof(template string, args ...interface{}) {
	l.Infof(template, args)
}
