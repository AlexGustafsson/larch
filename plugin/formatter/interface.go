package formatter

import (
	"io"
	"net/rpc"

	"github.com/AlexGustafsson/larch/warc"
	"github.com/hashicorp/go-plugin"
)

// FormatterName is the name of the formatter plugin.
const FormatterName = "formatter"

// HandshakeConfig is a common handshake that is shared by plugin and host.
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "FORMAT_PLUGIN",
	MagicCookieValue: "db7a94271e524a13a19e9264afa84b11",
}

// Formatter is a plugin implementing a new file format, such as WebArchive.
type Formatter interface {
	// Marshal a WARC file to the format specified by the plugin.
	Marshal(archive *warc.File) ([]byte, error)
	// Unmarshal a format specified by the plugin to a WARC file.
	Unmarshal(formattedArchive interface{}) (*warc.File, error)
	// Read reads a format specified by the plugin.
	Read(reader io.ReadSeeker) (interface{}, error)
	// Write writes a format specified by the plugin.
	Write(writer io.Writer, formattedArchive interface{}) error
}

// FormatterRPCClient is a RPC implementation of the plugin.
type FormatterRPCClient struct {
	client *rpc.Client
}

// FormatterRPCServer exposes the services consumed by PluginRPC.
type FormatterRPCServer struct {
	// Implementation is the real plugin implementation.
	Implementation Formatter
}

// FormatterPlugin ...
type FormatterPlugin struct {
	Implementation Formatter
}

// Server returns a server.
func (plugin *FormatterPlugin) Server(broker *plugin.MuxBroker) (interface{}, error) {
	return &FormatterRPCServer{Implementation: plugin.Implementation}, nil
}

// Client returns a client.
func (plugin *FormatterPlugin) Client(broker *plugin.MuxBroker, client *rpc.Client) (interface{}, error) {
	return &FormatterRPCClient{client: client}, nil
}
