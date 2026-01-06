package config

import (
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Sources    []Source            `yaml:"sources"`
	Strategies map[string]Strategy `yaml:"strategies"`
	Libraries  map[string]Library  `yaml:"libraries"`
}

type Source struct {
	Type        string   `yaml:"type,omitempty"`
	Name        string   `yaml:"name,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Strategy    string   `yaml:"strategy"`
	Options     *RawNode `yaml:"options,omitempty"`
}

type URLSourceOptions struct {
	URL string `yaml:"url"`
}

type FeedSourceOptions struct {
	URL      string        `yaml:"url"`
	Interval time.Duration `yaml:"interval"`
}

type Strategy struct {
	Name        string     `yaml:"name"`
	Description string     `yaml:"description"`
	Library     string     `yaml:"library"`
	Archivers   []Archiver `yaml:"archivers"`
}

type Archiver struct {
	Type    string   `yaml:"type"`
	Options *RawNode `yaml:"options,omitempty"`
}

type ChromeArchiverOptions struct {
	PDF        *ChromeArchiverPDFOptions        `yaml:"pdf,omitempty"`
	Singlefile *ChromeArchiverSinglefileOptions `yaml:"singlefile,omitempty"`
	Screenshot *ChromeArchiverScreenshotOptions `yaml:"screenshot,omitempty"`
}

type ChromeArchiverPDFOptions struct {
	Enabled bool `yaml:"enabled,omitempty"`
}

type ChromeArchiverSinglefileOptions struct {
	Enabled bool `yaml:"enabled,omitempty"`
}

type ChromeArchiverScreenshotOptions struct {
	Enabled     bool     `yaml:"enabled,omitempty"`
	Resolutions []string `yaml:"resolutions,omitempty"`
}

type Library struct {
	Type        string   `yaml:"type,omitempty"`
	Name        string   `yaml:"name,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Options     *RawNode `yaml:"options,omitempty"`
}

type DiskLibraryOptions struct {
	Path     string `yaml:"path"`
	ReadOnly bool   `yaml:"readOnly,omitempty"`
	Compress bool   `yaml:"compress,omitempty"`
}

type ArchiveBoxLibraryOptions struct {
	Path     string `yaml:"path"`
	ReadOnly bool   `yaml:"readOnly,omitempty"`
}

type RawNode struct{ node *yaml.Node }

func (n *RawNode) MarshalYAML() (any, error) {
	return n.node, nil
}

func (n *RawNode) UnmarshalYAML(node *yaml.Node) error {
	n.node = node
	return nil
}

func (n *RawNode) As(v any) error {
	return n.node.Decode(v)
}
