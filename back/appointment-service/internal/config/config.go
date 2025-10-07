package config

import (
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const envPrefix = "SHED_"
const envConfigPath = envPrefix + "CONFIG_PATH"

func LoadConfig() (*koanf.Koanf, error) {
	k := koanf.New(".")
	yamlPath := os.Getenv(envConfigPath)
	if yamlPath != "" {
		if err := k.Load(file.Provider(yamlPath), yaml.Parser()); err != nil {
			return nil, err
		}
	}

	if err := k.Load(env.Provider(".", env.Opt{
		Prefix: envPrefix,
		TransformFunc: func(k, v string) (string, any) {
			//SHED_SERVER__CONFIG_FILE --> server.config_file
			k = strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(k, envPrefix)), "__", ".")
			return k, v
		},
	}), nil); err != nil {
		return nil, err
	}

	return k, nil
}
