package main

import (
	"encoding/json"
	"os"
)

type MachineConfig struct {
	ImageConfig ImageConfig
	RootDevice  string
	TTY         bool
	Hostname    string
}

type ImageConfig struct {
	Cmd        []string
	Env        []string
	WorkingDir string
	User       string
}

func DecodeMachine(path string) (MachineConfig, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return MachineConfig{}, err
	}
	var config MachineConfig
	if err := json.Unmarshal(contents, &config); err != nil {
		return MachineConfig{}, err
	}
	return config, nil
}
