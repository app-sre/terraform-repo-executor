package pkg

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"gopkg.in/yaml.v2"
)

func processConfig(cfgPath string) (*Input, error) {
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	var cfg Input
	// internally yaml.Unmarshal will use json.Unmarshal if it detects json format
	err = yaml.Unmarshal(raw, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// generic function for executing commands on host
func executeCommand(dir, command string, args []string) (string, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stdout.String(), errors.New(stderr.String())
	}
	return stdout.String(), nil
}

func (r Repo) cloneRepo(workdir string) error {
	args := []string{"-c", fmt.Sprintf(
		// clone repo with specified name and checkout specified ref
		"git clone %s %s && cd %s && git checkout %s",
		r.URL,
		r.Name,
		r.Name,
		r.Ref,
	)}
	_, err := executeCommand(workdir, "/bin/sh", args)
	if err != nil {
		return err
	}
	return nil
}
