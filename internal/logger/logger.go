package logger

import (
	"go.uber.org/zap"
)

var log *zap.Logger

// Init initializes the global logger (called once at startup)
func Init(mode string) {
	var err error
	if mode == "production" {
		log, err = zap.NewProduction()
	} else {
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
