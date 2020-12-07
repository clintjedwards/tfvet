package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	ConfigPath string `split_words:"true" default:"~/.tfvet.d"`
}

// FromEnv parses environment variables into the config object based on envconfig name
func FromEnv() (*Config, error) {
	var config Config

	err := envconfig.Process("tfvet", &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
