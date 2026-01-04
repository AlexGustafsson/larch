package config

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRead(t *testing.T) {
	file, err := os.Open("testdata/config.yaml")
	require.NoError(t, err)
	defer file.Close()

	config, err := Read(file)
	assert.NoError(t, err)

	var diskOptions DiskLibraryOptions
	err = config.Libraries["disk"].Options.As(&diskOptions)
	require.NoError(t, err)
	fmt.Printf("%+v\n", diskOptions)

	fmt.Printf("%+v\n", config)

	v, _ := json.MarshalIndent(config, "", "  ")
	fmt.Printf("%s\n", v)
}
