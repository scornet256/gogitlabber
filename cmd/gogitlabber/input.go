package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/scornet256/go-logger"
	"gopkg.in/yaml.v3"
)

// config struct for config
type Config struct {
	Concurrency     int    `yaml:"concurrency"`
	Debug           bool   `yaml:"debug"`
	Destination     string `yaml:"destination"`
	GitBackend      string `yaml:"git_backend"`
	GitHost         string `yaml:"git_host"`
	GitToken        string `yaml:"git_token"`
	IncludeArchived string `yaml:"include_archived"`
}

// setdefaults sets default values for the configuration
func (conf *Config) setDefaults() {
	conf.Concurrency = 15
	conf.Debug = false
	conf.Destination = "$HOME/Documents"
	conf.GitBackend = ""
	conf.GitHost = "gitlab.com"
	conf.GitToken = ""
	conf.IncludeArchived = "excluded"
}

// expand variable paths
func expandPath(path string) string {

	// expand environment variables like $home
	expanded := os.ExpandEnv(path)

	// expand ~
	if strings.HasPrefix(expanded, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return expanded
		}
		expanded = filepath.Join(home, expanded[2:])
	} else if expanded == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return expanded
		}
		expanded = home
	}

	return filepath.Clean(expanded)
}

// loadconfig from yaml file
func loadConfig(configPath string) (*Config, error) {
	config := &Config{}
	config.setDefaults()

	// check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// read config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// parse yaml
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// validateConfig validates the configuration values
func (conf *Config) validateConfig() error {
	// validate required parameters
	if conf.GitToken == "" {
		return fmt.Errorf("git_token is required")
	}

	// validate archived option
	switch conf.IncludeArchived {
	case "any", "exclusive", "excluded":
	default:
		return fmt.Errorf("invalid include_archived option: %s (must be any|excluded|exclusive)", conf.IncludeArchived)
	}

	// validate concurrency
	if conf.Concurrency < 1 {
		return fmt.Errorf("concurrency must be greater than 0")
	}

	return nil
}

// process config after loading
func (conf *Config) processConfig() {
	// expand path variables
	conf.Destination = expandPath(conf.Destination)

	// add trailing slash if not provided
	if !strings.HasSuffix(conf.Destination, "/") {
		conf.Destination += "/"
	}
}

// log active config
func (conf *Config) logConfig(configPath string) {
	logger.Print("Configuration: Using config file: "+configPath, nil)
	logger.Print("Configuration: Using host: "+conf.GitHost, nil)
	logger.Print("Configuration: Using destination: "+conf.Destination, nil)
	logger.Print("Configuration: Using concurrency: "+fmt.Sprintf("%d", conf.Concurrency), nil)
	logger.Print("Configuration: Using archived option: "+conf.IncludeArchived, nil)
	if conf.Debug {
		logger.Print("Configuration: Debug mode enabled", nil)
	}
}

// manage arguments
func manageArguments() *Config {

	defaultConfigPath := "./$HOME./gogitlabber.yaml"

	// Define only the config file flag
	configFileFlag := flag.String(
		"config",
		defaultConfigPath,
		"Specify config file path (YAML)\n  example: -config=./config/app.yaml")

	versionFlag := flag.Bool("version", false, "Print the version and exit")

	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	configPath := *configFileFlag

	// Load configuration from YAML file
	config, err := loadConfig(configPath)
	if err != nil {
		flag.Usage()
		logger.Fatal("Configuration error: "+err.Error(), nil)
	}

	// Process configuration
	config.processConfig()

	// Validate configuration
	if err := config.validateConfig(); err != nil {
		flag.Usage()
		logger.Fatal("Configuration validation error: "+err.Error(), nil)
	}

	// Log configuration
	config.logConfig(configPath)

	return config
}
