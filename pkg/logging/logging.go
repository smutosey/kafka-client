package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

func InitLogger(logFilePath string) *logrus.Logger {
	log := logrus.New()

	// Логируем в файл и stdout
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.Out = file
	} else {
		logrus.Warn("Не удалось открыть файл лога, логирование в stdout")
		log.Out = os.Stdout
	}

	// Формат логирования
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	return log
}
