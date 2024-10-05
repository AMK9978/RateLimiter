package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Log is the global logger instance.
var Log *logrus.Logger // nolint

// Initialize sets up the logger with configurations.
func Initialize() {
	Log = logrus.New()

	Log.SetOutput(os.Stdout)

	Log.SetLevel(logrus.InfoLevel)

	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}
