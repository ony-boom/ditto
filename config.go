package main

import (
	"fmt"
	"log"
	"os"

	"github.com/adrg/xdg"
	"github.com/charmbracelet/log"
	"github.com/pelletier/go-toml/v2"
)

type ConfigFile struct {
	NoConfirm          *bool     `toml:"noConfirm"`
	AurHelper          *string   `toml:"aurHelper"`
	ExtraInstallArgs   *[]string `toml:"extraInstallArgs"`
	ExtraUninstallArgs *[]string `toml:"extraUninstallArgs"`
	UninstallIgnore    *[]string `toml:"uninstallIgnore"`
	Pager              *[]string `toml:"pager"`
}

type Config struct {
	NoConfirm          bool      `toml:"noConfirm"`
	AurHelper          string    `toml:"aurHelper"`
	ExtraInstallArgs   []string  `toml:"extraInstallArgs"`
	ExtraUninstallArgs []string  `toml:"extraUninstallArgs"`
	UninstallIgnore    []string  `toml:"uninstallIgnore"`
	Pager              *[]string `toml:"pager"`
}

const (
	configDoc = `# Ditto config
# noConfirm: add --noconfirm to pacman/aur commands
# aurHelper: name of the AUR helper to use (e.g. yay, paru)
# uninstallIgnore: list of packages that will never be uninstalled
# extraInstallArgs: additional arguments to pass to install commands
# extraUninstallArgs: additional arguments to pass to uninstall commands
# pager: command to use for displaying output with their arguments (e.g. less, bat)

`
	configPerm = 0644
)

var (
	defaultArgs   = []string{"-S"}
	defaultConfig = Config{
		NoConfirm:       true,
		AurHelper:       "",
		UninstallIgnore: nil,
	}
)

func getConfigPath() (string, error) {
	return xdg.ConfigFile("ditto/config.toml")
}

func parseConfigFile(data []byte) (*ConfigFile, error) {
	var raw ConfigFile
	return &raw, toml.Unmarshal(data, &raw)
}

func (cf *ConfigFile) toConfig() *Config {
	return &Config{
		NoConfirm:          ptrValueOrDefault(cf.NoConfirm, defaultConfig.NoConfirm),
		AurHelper:          ptrValueOrDefault(cf.AurHelper, defaultConfig.AurHelper),
		UninstallIgnore:    ptrValueOrDefault(cf.UninstallIgnore, defaultConfig.UninstallIgnore),
		ExtraInstallArgs:   ptrValueOrDefault(cf.ExtraInstallArgs, []string{}),
		ExtraUninstallArgs: ptrValueOrDefault(cf.ExtraUninstallArgs, []string{}),
		Pager:              cf.Pager,
	}
}

func LoadConfig() *Config {
	path, err := getConfigPath()
	if err != nil {
		log.Fatalln(fmt.Errorf("could not find config file path: %w", err))
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		saveConfig(path, &defaultConfig)
		return &defaultConfig
	}
	if err != nil {
		log.Fatalln(fmt.Errorf("could not read config: %w", err))
	}

	raw, err := parseConfigFile(data)
	if err != nil {
		saveConfig(path, &defaultConfig)
		return &defaultConfig
	}

	return raw.toConfig()
}

func saveConfig(path string, cfg *Config) {
	data, err := toml.Marshal(cfg)
	if err != nil {
		log.Fatalln(fmt.Errorf("could not marshal config: %w", err))
	}

	dataWithComments := append([]byte(configDoc), data...)
	if err := os.WriteFile(path, dataWithComments, configPerm); err != nil {
		log.Fatalln(fmt.Errorf("could not write config file: %w", err))
	}
}
