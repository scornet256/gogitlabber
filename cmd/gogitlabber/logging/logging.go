package logging

import (
	"log"
)

func LogPrint(debug bool, message string, err error) {
	if debug {
		if err != nil {
			log.Printf("gogitlabber | DEBUG: %v error: %v\n", message, err)
		}
		if err == nil {
			log.Printf("gogitlabber | DEBUG: %v\n", message)
		}
	}
}

func LogFatal(message string, err error) {
	if err != nil {
		log.Fatalf("gogitlabber | FATAL: %v error: %v\n", message, err)
	}
	log.Fatalf("gogitlabber | FATAL: %v\n", message)
}
