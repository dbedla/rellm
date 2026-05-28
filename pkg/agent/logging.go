package agent

import (
	"os"

	"github.com/rs/zerolog"
)

func NewBaseFileLogger(pathToFile string) zerolog.Logger {
	file, err := os.OpenFile(pathToFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}

	return zerolog.New(file).
		With().
		Timestamp().
		Logger()
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
