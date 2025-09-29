package logger

import (
	"sync"

	"go.uber.org/zap"
)

var once sync.Once
var logger *zap.Logger

func NewLogger(logLvl zap.AtomicLevel) (*zap.Logger, error) {
	var err error = nil
	// singleton
	once.Do(func() {
		config := zap.NewProductionConfig()
		config.Level = logLvl
		logger, err = config.Build()
		// logger = zap.NewNop()
		defer logger.Sync()
	})
	return logger, err
}
