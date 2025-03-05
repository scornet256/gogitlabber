package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
)

func manageArguments() {

	// configuration vars
	var archivedFlag = flag.String(
		"archived",
		"excluded",
		"To include archived repositories (any|excluded|exclusive)\n  example: -archived=any\nenv = GOGITLABBER_ARCHIVED\n")

	var concurrencyFlag = flag.Int(
		"concurrency",
		15,
		"Specify repository concurrency\n  example: -concurrency=15\nenv = GOGITLABBER_CONCURRENCY\n")

	var destinationFlag = flag.String(
		"destination",
		"$HOME/Documents",
		"Specify where to check the repositories out\n  example: -destination=$HOME/repos\nenv = GOGITLABBER_DESTINATION\n")

	var hostFlag = flag.String(
		"gitlab-url",
		"gitlab.com",
		"Specify GitLab host\n  example: -gitlab-url=gitlab.com\nenv = GITLAB_URL\n")

	var tokenFlag = flag.String(
		"gitlab-api-token",
		"",
		"Specify GitLab API token\n  example: -gitlab-api=glpat-xxxx\nenv = GITLAB_API_TOKEN\n")

	var debugFlag = flag.Bool(
		"debug",
		false,
		"Toggle debug mode\n example: -debug=true\nenv = GOGITLABBER_DEBUG\n")

	flag.Parse()

	// assign the parsed values to your variables
	concurrency = *concurrencyFlag
	gitlabHost = *hostFlag
	gitlabToken = *tokenFlag
	includeArchived = *archivedFlag
	repoDestinationPre = *destinationFlag
	debug = *debugFlag

	// manage verosity option
	switch envDebug := os.Getenv("GOGITLABBER_DEBUG"); {
	case envDebug != "":
		var err error
		debug, err = strconv.ParseBool(envDebug)
		logPrint("Configuration: debug option found", nil)
		if err != nil {
			logFatal("FATAL: config; not a valid bool", nil)
		}
	default:
		debug = false
	}

	// manage gitlab api option
	switch envToken := os.Getenv("GITLAB_API_TOKEN"); {
	case envToken != "":
		gitlabToken = envToken
		logPrint("Configuration: Gitlab API Token found", nil)
	default:
		flag.Usage()
		logFatal("Configuration: Giltab API Token not found", nil)
	}

	// manage gitlab url option
	switch envHost := os.Getenv("GITLAB_URL"); {
	case envHost != "":
		gitlabHost = envHost
		logPrint("Configuration: Gitlab host found", nil)
	default:
		flag.Usage()
		logFatal("Configuration: Gitlab host not found", nil)
	}

	// manage destination option
	switch envRepoDest := os.Getenv("GOGITLABBER_DESTINATION"); {
	case envRepoDest != "":
		repoDestinationPre = envRepoDest
		logPrint("Configuration: destination found", nil)
	default:
		flag.Usage()
		logFatal("Configuration: destination not found", nil)
	}

	// add slash 🎩🎸 if not provided
	switch {
	case !strings.HasSuffix(repoDestinationPre, "/"):
		repoDestinationPre += "/"
	}

	// manage concurrency option
	switch envConcurrency := os.Getenv("GOGITLABBER_CONCURRENCY"); {
	case envConcurrency == "":
		concurrency = 15
	case envConcurrency != "":
		concurrencyValue, err := strconv.Atoi(envConcurrency)
		if err != nil {
			logFatal("invalid concurrency value found in environment: %v", err)
		}
		concurrency = concurrencyValue
		logPrint("Configuration: concurrency option found", nil)
	default:
		flag.Usage()
		log.Fatalln("FATAL: config; concurrency not found")
	}

	// manage archived option
	switch envArchived := os.Getenv("GOGITLABBER_ARCHIVED"); {
	case envArchived == "":
		includeArchived = "excluded"
		logPrint("Configuration: archive option found", nil)

	case envArchived == "any":
		includeArchived = envArchived
		logPrint("Configuration: archive option found", nil)

	case envArchived == "exclusive":
		includeArchived = envArchived
		logPrint("Configuration: archive option found", nil)

	case envArchived == "excluded":
		includeArchived = envArchived
		logPrint("Configuration: archive option found", nil)

	default:
		flag.Usage()
		logFatal("FATAL: Configuration: no or wrong archive option found", nil)
	}
}
