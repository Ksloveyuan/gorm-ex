package gorm_ex

import (
	"log"
	"os"
)

type consoleLogger struct {
	info *log.Logger
	warn *log.Logger
	error *log.Logger
}

func newConsoleLogger() *consoleLogger  {
	return &consoleLogger{
		info:log.New(os.Stdout, "info", log.LstdFlags),
		warn:log.New(os.Stdout, "warn", log.LstdFlags),
		error:log.New(os.Stdout, "error", log.LstdFlags),
	}
}

func(cl *consoleLogger) LogInfoc(category, message string) {
	cl.info.Println( category, message)
}

func(cl *consoleLogger) LogWarnc(category string, err error, message string) {
	if err == nil{
		cl.warn.Println(category, message)
	} else {
		cl.warn.Println(category, err.Error(), message)
	}
}

func(cl *consoleLogger) LogErrorc(category string, err error, message string) {
	cl.error.Println(category, err.Error(), message)
}