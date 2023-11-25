package logger

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"runtime"
)

func RunCustomLogger(appName string) {
	filePath := getFolderLogPath(appName)
	lumberjackLogrotate := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    100, // megabytes
		MaxBackups: 50,
		MaxAge:     30,   //days
		Compress:   true, // disabled by default
	}

	logMultiWriter := io.MultiWriter(os.Stdout, lumberjackLogrotate)
	log.SetOutput(logMultiWriter)

}

func getFolderLogPath(appName string) string {
	switch runtime.GOOS {
	case "windows":
		userDirectory, _ := os.UserHomeDir()
		return fmt.Sprintf("\\%v\\AppData\\Roaming\\%v\\logs\\%v.log", userDirectory, appName, appName)
	case "darwin":
		return fmt.Sprintf("~/Library/Logs/%v/%v.log", appName, appName)
	default:
		return fmt.Sprintf("./logs/%v.log", appName)
	}
}
