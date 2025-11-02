package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type (
	DaemonConfiguration struct {
		Server   ServerConfiguration   `json:"server"`
		Database DatabaseConfiguration `json:"database"`
	}

	ServerConfiguration struct {
		Address string `json:"addr"`
		Port    uint16 `json:"port"`
	}

	DatabaseConfiguration struct {
		Path string `json:"path"`
	}
)

func Load(path string) (DaemonConfiguration, error) {
	path = filepath.Join(path, "config.json")

	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), nil
		}
		return DaemonConfiguration{}, fmt.Errorf("failed to load configuration file (%s): %s", path, err)
	}
	defer f.Close()

	var c DaemonConfiguration
	d := json.NewDecoder(f)
	if err := d.Decode(&c); err != nil {
		return DaemonConfiguration{}, fmt.Errorf("failed to read configuration file (%s): %s", path, err)
	}

	return fillDefault(c), nil
}

func Default() DaemonConfiguration {
	return DaemonConfiguration{
		Server: ServerConfiguration{
			Address: "127.0.0.1",
			Port:    25697,
		},
		Database: DatabaseConfiguration{
			Path: "/var/lib/mirror-sync/data.db",
		},
	}
}

func fillDefault(c DaemonConfiguration) DaemonConfiguration {
	if len(c.Database.Path) == 0 {
		c.Database.Path = Default().Database.Path
	}
	if len(c.Server.Address) == 0 {
		c.Server.Address = Default().Server.Address
	}
	if c.Server.Port == 0 {
		c.Server.Port = Default().Server.Port
	}
	return c
}
