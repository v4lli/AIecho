package logging

import (
	"log"
	"os"
)

var logFile *os.File

func SetupLogging() {
	logFile, err := os.OpenFile("pipeline.log", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Printf("error opening log file: %v", err)
	}

	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
}

func CloseLogging() {
	if logFile != nil {
		if err := logFile.Close(); err != nil {
			log.Printf("error closing log file: %v", err)
		}
	}
}
