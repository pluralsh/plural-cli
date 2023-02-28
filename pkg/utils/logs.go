package utils

import (
	"io"
	"log"
	"os"
)

func init() {
	EnableDebug = false
}

var infoLogger *log.Logger
var errorLogger *log.Logger

var EnableDebug bool

func LogInfo() *log.Logger {
	if infoLogger == nil {
		infoLogger = log.New(getOutputWriter(), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	}
	return infoLogger
}

func LogError() *log.Logger {
	if errorLogger == nil {
		errorLogger = log.New(getOutputWriter(), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	}
	return errorLogger
}

func getOutputWriter() (out io.Writer) {
	out = os.Stdout
	if !EnableDebug {
		out = io.Discard
	}
	return
}
