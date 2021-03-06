package providers

import (
	// "os/exec"
	"github.com/duolacloud/microbase/config"

	// "go.uber.org/zap"
	// "go.uber.org/zap/zapcore"
	"github.com/micro/go-micro/v2/logger"
	_zap "github.com/micro/go-plugins/logger/zap/v2"
)

func InitLogger(config *config.AppConfig) error {
	l, err := _zap.NewLogger()
	if err != nil {
		return err
	}

	logger.DefaultLogger = l

	logOptions := make([]logger.Option, 0)
	if config.Env == "dev" {
		logOptions = append(logOptions, logger.WithLevel(logger.DebugLevel))
	}

	logger.Init(logOptions...)

	return nil
}
