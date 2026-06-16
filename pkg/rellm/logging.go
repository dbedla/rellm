package rellm

import (
	"os"

	"github.com/rs/zerolog"
)

func NewBaseFileLogger(pathToFile string) (zerolog.Logger, error) {
	file, err := os.OpenFile(pathToFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return zerolog.Logger{}, err
	}

	return zerolog.New(file).
		With().
		Timestamp().
		Logger(), nil
}

func NewComponentLogger(base zerolog.Logger, component string) zerolog.Logger {
	return base.With().
		Str("component", component).
		Logger()
}

func NewBaseStdOutLogger() zerolog.Logger {
	multi := zerolog.MultiLevelWriter(os.Stdout)
	l := zerolog.New(multi).With().Timestamp().Logger()
	return l
}
