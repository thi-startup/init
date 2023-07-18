package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/creack/pty"
	"github.com/sirupsen/logrus"
)

type Process struct {
	args       []string
	env        []string
	workingDir string
}

func NewProcess(cfg MachineConfig) (*Process, error) {
	if len(cfg.ImageConfig.Cmd) < 1 && len(cfg.CmdOverride) < 1 {
		return nil, fmt.Errorf("error no cmd provided")
	}

	args := append(cfg.ImageConfig.Entrypoint, cfg.ImageConfig.Cmd...)
	if len(cfg.CmdOverride) > 0 {
		args = cfg.CmdOverride
	}

	envars := append(cfg.ImageConfig.Env, cfg.ExtraEnv...)

	if err := populateProcessEnv(envars); err != nil {
		return nil, fmt.Errorf("error populating process env: %v", err)
	}

	return &Process{
		args:       args,
		env:        envars,
		workingDir: cfg.ImageConfig.WorkingDir,
	}, nil
}

func (p *Process) Run() (*os.File, error) {
	lp, err := exec.LookPath(p.args[0])
	if err != nil {
		return nil, fmt.Errorf("error searching for process: %v", err)
	}

	cmd := &exec.Cmd{
		Path: lp,
		Args: p.args,
		Env:  p.env,
		Dir:  p.workingDir,
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logrus.Infof("Running %s", cmd.String())

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return ptmx, err
}

func parseCmdArgs(cmd []string) (string, []string, error) {
	cmdPath, err := exec.LookPath(cmd[0])
	if err != nil {
		return "", nil, err
	}

	if len(cmd) == 1 {
		return cmdPath, nil, nil
	}
	return cmdPath, cmd[1:], nil
}

func populateProcessEnv(env []string) error {
	for _, pair := range env {
		p := strings.SplitN(pair, "=", 2)
		if len(p) < 2 {
			return errors.New("invalid env var: missing '='")
		}
		name, val := p[0], p[1]
		if name == "" {
			return errors.New("invalid env var: name cannot be empty")
		}
		if strings.IndexByte(name, 0) >= 0 {
			return errors.New("invalid env var: name contains null byte")
		}
		if strings.IndexByte(val, 0) >= 0 {
			return errors.New("invalid env var: value contains null byte")
		}
		if err := os.Setenv(name, val); err != nil {
			return fmt.Errorf("could not set env var: system shit: %v", err)
		}
	}
	return nil
}
