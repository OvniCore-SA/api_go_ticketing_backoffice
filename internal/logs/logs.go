package logs

import (
	"log"
	"os"
)

var (
	infoLogger  = log.New(os.Stdout, "[INFO]  ", log.Ldate|log.Ltime)
	errorLogger = log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime)
	fatalLogger = log.New(os.Stderr, "[FATAL] ", log.Ldate|log.Ltime)
)

func Info(msg string) {
	infoLogger.Println(msg)
}

func Error(msg string) {
	errorLogger.Println(msg)
}

func Fatal(msg string) {
	fatalLogger.Fatalln(msg)
}
