package config

import (
	"io"
	"os"

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

func ReadFile(name string) (*Config, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return Read(file)
}
