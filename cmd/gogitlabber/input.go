package main

import (
	"flag"
	"gogitlabber/cmd/gogitlabber/logging"
	"os"
	"strconv"
	"strings"
)

// set default values and override values from environment variables
func setDefaultsFromEnv() {

	// set default values
	debug = false
	concurrency = 15
	gitlabHost = "gitlab.com"
	gitlabToken = ""
	includeArchived = "excluded"
	repoDestinationPre = "$HOME/Documents"

	// override with environment variables if present
	if envDebug := os.Getenv("GOGITLABBER_DEBUG"); envDebug != "" {
		if debugVal, err := strconv.ParseBool(envDebug); err == nil {
			debug = debugVal
		} else {
			logging.LogPrint(debug, "Warning: Invalid debug value in environment, using default", nil)
		}
	}

	if envToken := os.Getenv("GITLAB_API_TOKEN"); envToken != "" {
		gitlabToken = envToken
	}

	if envHost := os.Getenv("GITLAB_URL"); envHost != "" {
		gitlabHost = envHost
	}

	if envRepoDest := os.Getenv("GOGITLABBER_DESTINATION"); envRepoDest != "" {
		repoDestinationPre = envRepoDest
	}

	if envConcurrency := os.Getenv("GOGITLABBER_CONCURRENCY"); envConcurrency != "" {
		if concurrencyVal, err := strconv.Atoi(envConcurrency); err == nil {
			concurrency = concurrencyVal
		} else {
			logging.LogPrint(debug, "Warning: Invalid concurrency value in environment, using default", nil)
		}
	}

	if envArchived := os.Getenv("GOGITLABBER_ARCHIVED"); envArchived != "" {
		switch envArchived {
		case "any", "exclusive", "excluded":
			includeArchived = envArchived
		default:
			logging.LogPrint(debug, "Warning: Invalid archived value in environment, using default", nil)
		}
	}
}

func manageArguments() {

	// set defaults from environment variables
	setDefaultsFromEnv()

	// define flags (which will override environment variables)
	var archivedFlag = flag.String(
		"archived",
		includeArchived,
		"To include archived repositories (any|excluded|exclusive)\n  example: -archived=any\nenv = GOGITLABBER_ARCHIVED\n")

	var concurrencyFlag = flag.Int(
		"concurrency",
		concurrency,
		"Specify repository concurrency\n  example: -concurrency=15\nenv = GOGITLABBER_CONCURRENCY\n")

	var destinationFlag = flag.String(
		"destination",
		repoDestinationPre,
		"Specify where to check the repositories out\n  example: -destination=$HOME/repos\nenv = GOGITLABBER_DESTINATION\n")

	var hostFlag = flag.String(
		"gitlab-url",
		gitlabHost,
		"Specify GitLab host\n  example: -gitlab-url=gitlab.com\nenv = GITLAB_URL\n")

	var tokenFlag = flag.String(
		"gitlab-api-token",
		gitlabToken,
		"Specify GitLab API token\n  example: -gitlab-api=glpat-xxxx\nenv = GITLAB_API_TOKEN\n")

	var debugFlag = flag.Bool(
		"debug",
		debug,
		"Toggle debug mode\n  example: -debug=true\nenv = GOGITLABBER_DEBUG\n")

	flag.Parse()

	// override with flag values (higher precedence)
	concurrency = *concurrencyFlag
	debug = *debugFlag
	gitlabHost = *hostFlag
	gitlabToken = *tokenFlag
	includeArchived = *archivedFlag
	repoDestinationPre = *destinationFlag

	// add slash 🎩🎸 if not provided
	if !strings.HasSuffix(repoDestinationPre, "/") {
		repoDestinationPre += "/"
	}

	// validate required parameters
	if gitlabToken == "" {
		flag.Usage()
		logging.LogFatal("Configuration: Gitlab API Token not found", nil)
	}

	// validate archived option
	switch includeArchived {
	case "any", "exclusive", "excluded":
	default:
		flag.Usage()
		logging.LogFatal("Configuration: Invalid archive option: "+includeArchived, nil)
	}

	// log configuration
	logging.LogPrint(debug, "Configuration: Using GitLab host: "+gitlabHost, nil)
	logging.LogPrint(debug, "Configuration: Using destination: "+repoDestinationPre, nil)
	logging.LogPrint(debug, "Configuration: Using concurrency: "+strconv.Itoa(concurrency), nil)
	logging.LogPrint(debug, "Configuration: Using archived option: "+includeArchived, nil)
	if debug {
		logging.LogPrint(debug, "Configuration: Debug mode enabled", nil)
	}
}
