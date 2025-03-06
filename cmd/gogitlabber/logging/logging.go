package logging

import (
	"log"
)

// Prints the formatted log, taking both a message (string) and optionally
// an error as inputs.
func Print(message string, err error) {
	if debug {
		log.Printf(applicationName+" | LOG: %v\n", message)
		if err != nil {
			log.Printf(applicationName+" | ERROR: %v\n", err)
		}
	}
}

// Prints the fatal error and exits the application. Takes both the message and
// optionally an error as inputs.
func Fatal(message string, err error) {
	log.Fatalf(applicationName+" | FATAL: %v\n", message)
	if err != nil {
		log.Fatalf(applicationName+" | ERROR: %v\n", err)
	}
}
