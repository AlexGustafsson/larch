package plugin

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/AlexGustafsson/larch/plugin/formatter"
	"github.com/hashicorp/go-plugin"
)

type Manager struct {
	Formatters map[string]*formatter.Formatter
	clients    []*plugin.Client
}

// NewManager creates a new manager. Note, always defer manager.Kill()!
func NewManager() *Manager {
	return &Manager{
		Formatters: make(map[string]*formatter.Formatter, 0),
		clients:    make([]*plugin.Client, 0),
	}
}

// RegisterFormatter registers a formatter plugin based on a path.
func (manager *Manager) RegisterFormatter(name string, path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("Unable to access plugin: %v", err)
	}

	pluginMap := map[string]plugin.Plugin{
		"formatter": &formatter.FormatterPlugin{},
	}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: formatter.HandshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command(path),
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return fmt.Errorf("Unable to connect to plugin: %v", err)
	}

	raw, err := rpcClient.Dispense(formatter.FormatterName)
	if err != nil {
		client.Kill()
		return fmt.Errorf("Unable to dispense plugin: %v", err)
	}

	formatterPlugin := raw.(formatter.Formatter)
	manager.Formatters[name] = &formatterPlugin
	manager.clients = append(manager.clients, client)

	return nil
}

// Kill kills all spawned plugins.
func (manager *Manager) Kill() {
	for _, client := range manager.clients {
		client.Kill()
	}
}
