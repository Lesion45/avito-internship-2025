package logger

import "go.uber.org/zap"

const (
	logDev  = "dev"
	logProd = "prod"
)

// NewZap returns a new instance of the zap.Logger.
//
// newLogger creates a zap.Logger based on the "ENV" variable.
// Uses zap.NewDevelopment() for "dev" and zap.NewProduction() for "prod".
// Defaults to zap.NewProduction() for invalid or missing "ENV" values.
func NewZap(logLevel string) *zap.Logger {
	var log *zap.Logger

	switch logLevel {
	case logDev:
		log = zap.Must(zap.NewDevelopment())
	case logProd:
		log = zap.Must(zap.NewProduction())
	default:
		log = zap.Must(zap.NewDevelopment())
	}

	return log
}
