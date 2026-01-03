package config

import (
	"io"

	"gopkg.in/yaml.v3"
)

func Read(r io.Reader) (*Config, error) {
	var config Config

	decoder := yaml.NewDecoder(r)
	decoder.KnownFields(true)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
