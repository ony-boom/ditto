package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Pacman (or AUR helper) wrapper
type Pacman struct {
	binary             string
	isAurHelper        bool
	noConfirm          bool
	extraInstallArgs   []string
	extraUninstallArgs []string
}

// NewPacman creates a new Pacman wrapper based on Config
func NewPacman(cfg *Config) *Pacman {
	bin := "pacman"
	if cfg.AurHelper != "" {
		bin = cfg.AurHelper
	}

	return &Pacman{
		binary:             bin,
		isAurHelper:        cfg.AurHelper != "",
		noConfirm:          cfg.NoConfirm,
		extraInstallArgs:   cfg.ExtraInstallArgs,
		extraUninstallArgs: cfg.ExtraUninstallArgs,
	}
}

// exec builds an exec.Cmd with sudo when needed
func (p *Pacman) exec(args []string) *exec.Cmd {
	if p.isAurHelper {
		return exec.Command(p.binary, args...)
	}
	// Pacman requires sudo
	baseArgs := append([]string{p.binary}, args...)
	return exec.Command("sudo", baseArgs...)
}

// ListInstalled returns all installed package names
func (p *Pacman) ListInstalled() ([]string, error) {
	cmd := exec.Command(p.binary, "-Q")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list installed packages: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	pkgs := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// pacman -Q output: "pkgname version"
		parts := strings.SplitN(line, " ", 2)
		pkgs = append(pkgs, parts[0])
	}

	return pkgs, nil
}

// Install packages with optional extra args
func (p *Pacman) Install(pkgs []string, extraArgs ...string) error {
	args := []string{"-S"}
	args = append(args, extraArgs...)
	args = append(args, p.extraInstallArgs...)

	if p.noConfirm {
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

// Remove packages with optional extra args
func (p *Pacman) Remove(pkgs []string, extraArgs ...string) error {
	args := []string{"-R"}
	args = append(args, extraArgs...)
	args = append(args, p.extraUninstallArgs...)
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
