package logger

import (
	log "github.com/sirupsen/logrus"
	"os"
)

var Logger *log.Logger

func SetLogLevel(level log.Level) {
	Logger = log.New()
	Logger.Out = os.Stdout
	Logger.SetLevel(level)
	Logger.Formatter = &log.TextFormatter{
		ForceQuote:      true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	}
}
