package logging

import (
	"fmt"
	"io"
	"log"
)

// define application name used in prefix
var applicationName = ""

func SetAppName(name string) {
	applicationName = name
}

func GetAppName() string {
	return applicationName
}

var debug bool

func SetDebug(debugSetting bool) {
	debug = debugSetting
	if !debug {
		log.SetOutput(io.Discard)
	}
}

func GetDebug() bool {
	return debug
}

// Prints the formatted log, taking both a message (string) and optionally
// an error as inputs.
func Print(message string, err error) error {
	if debug {
		if err != nil {
			log.Printf(applicationName+" | DEBUG: %v error: %v\n", message, err)
		}
		if err == nil {
			log.Printf(applicationName+" | DEBUG: %v\n", message)
		}
	} else {
		return fmt.Errorf("It seems you want to print a log while debug is off")
	}
	return nil
}

// Prints the fatal error and exits the application. Takes both the message and
// optionally an error as inputs.
func Fatal(message string, err error) {
	if err != nil {
		log.Fatalf(applicationName+" | FATAL: %v error: %v\n", message, err)
	}
	log.Fatalf(applicationName+" | FATAL: %v\n", message)
}
