package logger

import (
	"go.uber.org/zap"
)

var log *zap.Logger

func Init(mode string) {
	var err error
	switch mode {
	case "production":
		log, err = zap.NewProduction()
	case "testing":
		log = zap.NewNop()
	default:
		log, err = zap.NewDevelopment()
	}

	if err != nil {
		panic(err)
	}
}

func L() *zap.Logger {
	if log == nil {
		panic("logger not initialized — call logger.Init() first")
	}
	return log
}
