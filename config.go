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
	Mounts      []Mounts
	EtcResolv   EtcResolv
	EtcHost     []EtcHost
}

type ImageConfig struct {
	Cmd        []string
	Entrypoint []string
	Env        []string
	WorkingDir string
	User       string
}

type Mounts struct {
	MountPath  string
	DevicePath string
}

type EtcResolv struct {
	Nameservers []string
}

type EtcHost struct {
	Host string
	IP   string
	Desc string
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
