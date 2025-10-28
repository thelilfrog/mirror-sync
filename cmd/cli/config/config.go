package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type (
	ClientConfiguration struct {
		Deamon ClientDaemonConfiguration `json:"daemon"`
	}

	ClientDaemonConfiguration struct {
		URL string `json:"url"`
	}
)

func Load() ClientConfiguration {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		panic("failed to get user config path: " + err.Error())
	}

	path := filepath.Join(userConfigDir, "mirror-sync", "config.json")

	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default()
		}
		panic("failed to load configuration file (" + path + "): " + err.Error())
	}
	defer f.Close()

	var c ClientConfiguration
	d := json.NewDecoder(f)
	if err := d.Decode(&c); err != nil {
		panic("failed to read the configuration file: " + err.Error())
	}

	return c
}

func Default() ClientConfiguration {
	return ClientConfiguration{
		Deamon: ClientDaemonConfiguration{
			URL: "http://localhost:25697",
		},
	}
}
