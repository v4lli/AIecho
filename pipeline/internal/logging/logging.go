package logging

import (
	"log"
	"os"
)

func SetupLogging() {
	logfile, err := os.OpenFile("pipeline.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("error opening log file: %v", err)
	}
	defer func() {
		if err := logfile.Close(); err != nil {
			log.Printf("error closing log file: %v", err)
		}
	}()

	log.SetOutput(logfile)
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
}
