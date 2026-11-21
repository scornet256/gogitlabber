package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/scornet256/go-logger"
	"gopkg.in/yaml.v3"
)

const defaultConfigPath = "~/.config/gogitlabber.yaml"

// config struct for config
type Config struct {
	Concurrency     int    `yaml:"concurrency"`
	Debug           bool   `yaml:"debug"`
	Destination     string `yaml:"destination"`
	GitBackend      string `yaml:"git_backend"`
	GitHost         string `yaml:"git_host"`
	GitProtocol     string `yaml:"git_protocol"`
	GitSSHPort      int    `yaml:"git_ssh_port"`
	GitSSHUser      string `yaml:"git_ssh_user"`
	GitToken        string `yaml:"git_token"`
	GitUserMail     string `yaml:"git_user_mail"`
	GitUserName     string `yaml:"git_user_name"`
	IncludeArchived string `yaml:"include_archived"`
}

// default configuration values
func defaultConfig() *Config {
	return &Config{
		Concurrency:     15,
		Destination:     "$HOME/Documents",
		GitHost:         "gitlab.com",
		GitProtocol:     "https",
		GitSSHPort:      22,
		GitSSHUser:      "git",
		IncludeArchived: "excluded",
	}
}

// expand environment variables and ~ in paths
func expandPath(path string) string {
	expanded := os.ExpandEnv(path)

	if expanded == "~" || strings.HasPrefix(expanded, "~/") {
		if homeDir, err := os.UserHomeDir(); err == nil {
			expanded = filepath.Join(homeDir, expanded[1:])
		}
	}

	return filepath.Clean(expanded)
}

// load config from yaml file; unknown keys are rejected to catch typos
func loadConfig(configPath string) (*Config, error) {
	cfg := defaultConfig()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(cfg); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("parsing config file %s: %w", configPath, err)
	}

	return cfg, nil
}

// validate the configuration values
func (conf *Config) validateConfig() error {
	switch conf.GitBackend {
	case "gitlab", "gitea":
	case "":
		return errors.New("git_backend is required (gitlab|gitea)")
	default:
		return fmt.Errorf("unsupported git_backend: %s (must be gitlab|gitea)", conf.GitBackend)
	}

	if conf.GitHost == "" {
		return errors.New("git_host is required")
	}

	if conf.GitToken == "" {
		return errors.New("git_token is required")
	}

	switch conf.GitProtocol {
	case "https":
	case "ssh":
		if conf.GitSSHPort < 1 || conf.GitSSHPort > 65535 {
			return fmt.Errorf("invalid git_ssh_port: %d (must be 1-65535)", conf.GitSSHPort)
		}
		if conf.GitSSHUser == "" {
			return errors.New("git_ssh_user is required for git_protocol ssh")
		}
	default:
		return fmt.Errorf("invalid git_protocol: %s (must be https|ssh)", conf.GitProtocol)
	}

	switch conf.IncludeArchived {
	case "any", "excluded", "exclusive":
	default:
		return fmt.Errorf("invalid include_archived option: %s (must be any|excluded|exclusive)", conf.IncludeArchived)
	}

	if conf.Concurrency < 1 {
		return errors.New("concurrency must be greater than 0")
	}

	return nil
}

// normalize config after loading
func (conf *Config) processConfig() {
	conf.Destination = expandPath(conf.Destination)

	// accept git_host with scheme or trailing slash
	conf.GitHost = strings.TrimPrefix(conf.GitHost, "https://")
	conf.GitHost = strings.TrimPrefix(conf.GitHost, "http://")
	conf.GitHost = strings.TrimSuffix(conf.GitHost, "/")
}

// log active config
func (conf *Config) logConfig(configPath string) {
	logger.Print("Configuration: Using config file: "+configPath, nil)
	logger.Print("Configuration: Using backend: "+conf.GitBackend, nil)
	logger.Print("Configuration: Using host: "+conf.GitHost, nil)
	logger.Print("Configuration: Using protocol: "+conf.GitProtocol, nil)
	logger.Print("Configuration: Using destination: "+conf.Destination, nil)
	logger.Print(fmt.Sprintf("Configuration: Using concurrency: %d", conf.Concurrency), nil)
	logger.Print("Configuration: Using archived option: "+conf.IncludeArchived, nil)
	if conf.Debug {
		logger.Print("Configuration: Debug mode enabled", nil)
	}
}

// manage arguments
func manageArguments() *Config {
	configFileFlag := flag.String(
		"config",
		defaultConfigPath,
		"Specify config file path (YAML)\n  example: -config=./config/app.yaml")
	versionFlag := flag.Bool("version", false, "Print the version and exit")
	debugFlag := flag.Bool("debug", false, "Enable debug mode")

	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	configPath := expandPath(*configFileFlag)

	cfg, err := loadConfig(configPath)
	if err != nil {
		flag.Usage()
		fatal("configuration error", err)
	}

	// override debug setting if flag is set
	if *debugFlag {
		cfg.Debug = true
	}

	cfg.processConfig()

	if err := cfg.validateConfig(); err != nil {
		flag.Usage()
		fatal("configuration validation error", err)
	}

	cfg.logConfig(configPath)

	return cfg
}
