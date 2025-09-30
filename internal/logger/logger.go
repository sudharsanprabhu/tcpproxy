package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type logFormatter struct {}

func (f *logFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	message := fmt.Sprintf("%s [%s] %s\n", entry.Time.Format("2006-01-02 15:04:05"), strings.ToUpper(entry.Level.String()), entry.Message)
	return []byte(message), nil
}


func GetLogger(name string) *logrus.Logger {
	lumber := &lumberjack.Logger{
		Filename: fmt.Sprintf("logs/%s.log", name),
		MaxAge: 7,
	}

	go func() {
		for {
			<-time.After(24 * time.Hour)
			lumber.Rotate()
		}
	}()

	logger := logrus.New()
	logger.SetFormatter(&logFormatter{})
	logger.Out = io.MultiWriter(os.Stdout, lumber)
	return logger
}