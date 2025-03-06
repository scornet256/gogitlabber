package logging

import (
	"log"
)

func Print(debug bool, message string, err error) {
	if debug {
		if err != nil {
			log.Printf("gogitlabber | DEBUG: %v error: %v\n", message, err)
		}
		if err == nil {
			log.Printf("gogitlabber | DEBUG: %v\n", message)
		}
	}
}

func Fatal(message string, err error) {
	if err != nil {
		log.Fatalf("gogitlabber | FATAL: %v error: %v\n", message, err)
	}
	log.Fatalf("gogitlabber | FATAL: %v\n", message)
}
