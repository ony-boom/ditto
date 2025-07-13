package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Pacman struct {
	command string
}

func NewPacman(cfg *Config) *Pacman {
	p := Pacman{
		command: "pacman",
	}

	if cfg.AurHelper != "" {
		p.command = cfg.AurHelper
	}

	return &p
}

func (p *Pacman) ListInstalled() ([]string, error) {
	cmd := exec.Command(p.command, "-Q")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list installed packages: %w", err)
	}

	installedPackages := []string{}

	for line := range strings.SplitSeq(string(out), "\n") {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		packageName := strings.SplitN(line, " ", 2)[0]
		installedPackages = append(installedPackages, packageName)
	}

	return installedPackages, nil
}

func (p *Pacman) isAurHelper() bool {
	return p.command != "pacman"
}

func (p *Pacman) Install(pkgs []string) error {
	args := []string{"-S"}

	if cfg.NoConfirm {
		args = append(args, "--noconfirm")
	}

	args = append(args, pkgs...)

	var cmd *exec.Cmd
	if p.isAurHelper() {
		cmd = exec.Command(p.command, args...)
	} else {
		fullArgs := append([]string{p.command}, args...)
		cmd = exec.Command("sudo", fullArgs...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install packages: %w", err)
	}

	return nil
}
