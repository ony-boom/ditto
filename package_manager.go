package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Pacman struct {
	commandStr string
}

func NewPacman(cfg *Config) *Pacman {
	p := Pacman{
		commandStr: "pacman",
	}

	if cfg.AurHelper != "" {
		p.commandStr = cfg.AurHelper
	}

	return &p
}

func (p *Pacman) exec(args []string) *exec.Cmd {
	if p.isAurHelper() {
		return exec.Command(p.commandStr, args...)
	}

	baseArgs := []string{p.commandStr}
	baseArgs = append(baseArgs, args...)
	return exec.Command("sudo", baseArgs...)
}

func (p *Pacman) ListInstalled() ([]string, error) {
	cmd := exec.Command(p.commandStr, "-Q")
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
	return p.commandStr != "pacman"
}

func (p *Pacman) Install(pkgs []string) error {
	args := []string{"-S"}

	if cfg.NoConfirm {
		args = append(args, "--noconfirm")
	}

	args = append(args, pkgs...)

	cmd := p.exec(args)

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install packages: %w", err)
	}

	return nil
}

func (p *Pacman) Remove(pkgs []string) error {
	args := []string{"-R"}

	args = append(args, pkgs...)

	cmd := p.exec(args)

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove packages: %w", err)
	}

	return nil
}
