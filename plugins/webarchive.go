package main

import (
	"io"

	"github.com/AlexGustafsson/larch/plugin/formatter"
	"github.com/AlexGustafsson/larch/plugins/webarchive"
	"github.com/AlexGustafsson/larch/warc"
	"github.com/hashicorp/go-plugin"
)

// WebArchiveFormatter is a formatter implementation for WebArchives
type WebArchiveFormatter struct{}

// Marshal ...
func (formatter *WebArchiveFormatter) Marshal(archive *warc.File) ([]byte, error) {
	return nil, nil
}

// Unmarshal ...
func (formatter *WebArchiveFormatter) Unmarshal(formattedArchive interface{}) (*warc.File, error) {
	return nil, nil
}

// Read ...
func (formatter *WebArchiveFormatter) Read(reader io.ReadSeeker) (interface{}, error) {
	return webarchive.Read(reader)
}

// Write ...
func (formatter *WebArchiveFormatter) Write(writer io.Writer, formattedArchive interface{}) error {
	file := formattedArchive.(webarchive.WebArchive)
	return file.Write(writer)
}

func main() {
	implementation := &WebArchiveFormatter{}

	pluginMap := map[string]plugin.Plugin{
		formatter.FormatterName: &formatter.FormatterPlugin{Implementation: implementation},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: formatter.HandshakeConfig,
		Plugins:         pluginMap,
	})
}
