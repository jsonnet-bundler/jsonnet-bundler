package pkg

import (
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const RegistryFile = "~/.config/jsonnet-bundler/registries.yaml"

var Registries *RegistryConfig

type RegistryConfig struct {
	// Dir is the path where the registry config(s) are stored
	Entries []GitRegistry `yaml:"registries"`
}

func DefaultConfig() *RegistryConfig {
	return &RegistryConfig{
		Entries: []GitRegistry{
			{
				Name:        "Default",
				Description: "The default registry.",
				Source:      "https://github.com/dadav/jb-registry-example.git",
				PackageFile: "_gen.json",
			},
		},
	}
}

func expandTilde(path string) (string, error) {
	if len(path) > 0 && path[0] == '~' {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		return filepath.Join(usr.HomeDir, path[1:]), nil
	}
	return path, nil
}

func (c RegistryConfig) SaveRegistries() error {
	confFile, err := expandTilde(RegistryFile)
	if err != nil {
		return err
	}
	yamlData, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(confFile), os.ModePerm)
	if err != nil {
		return err
	}
	err = os.WriteFile(confFile, yamlData, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func LoadRegistries() error {
	confFile, err := expandTilde(RegistryFile)
	if err != nil {
		return err
	}
	fileContent, err := os.ReadFile(confFile)
	if err != nil {
		Registries = DefaultConfig()
		err := Registries.SaveRegistries()
		if err != nil {
			return err
		}
		return nil
	}

	err = yaml.Unmarshal(fileContent, &Registries)
	if err != nil {
		return err
	}

	return nil
}
