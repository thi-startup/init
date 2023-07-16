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
	cmd *exec.Cmd
}

func NewProcess(cfg ImageConfig) (*Process, error) {
	if len(cfg.Cmd) < 1 {
		return nil, errors.New("no command to execute")
	}

	args := append(cfg.Entrypoint, cfg.Cmd...)

	name, args, err := parseCmdArgs(args)
	if err != nil {
		return nil, err
	}

	c := &exec.Cmd{
		Path: name,
		Args: append([]string{name}, args...),
		Env:  cfg.Env,
		Dir:  cfg.WorkingDir,
	}

	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return &Process{cmd: c}, nil
}

func (p *Process) Run() (*os.File, error) {
	logrus.Infof("Running %s", p.cmd.String())

	ptmx, err := pty.Start(p.cmd)
	if err != nil {
		return nil, err
	}

	if err := p.cmd.Wait(); err != nil {
		return nil, err
	}

	return ptmx, err
}

func (p *Process) Wait() error {
	return p.cmd.Wait()
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
