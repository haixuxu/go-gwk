package utils

import (
	"github.com/bbk47/toolbox"
	"os"
)

func NewLogger(label, level string) *toolbox.Logger {
	logger := toolbox.Log.NewLogger(os.Stdout, label)
	logger.SetLevel(level)
	return logger
}
