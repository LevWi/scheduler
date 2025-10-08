package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const envPrefix = "SCHED_"
const envConfigPath = envPrefix + "CONFIG_PATH"

func LoadConfigRaw() (*koanf.Koanf, error) {
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
			//SCHED_SERVER__CONFIG_FILE --> server.config_file
			k = strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(k, envPrefix)), "__", ".")
			return k, v
		},
	}), nil); err != nil {
		return nil, err
	}

	return k, nil
}

func configErr(err error) error {
	return fmt.Errorf("config err: %w", err)
}

func LoadConfigTo(tag string, out any) error {
	k, err := LoadConfigRaw()
	if err != nil {
		return configErr(err)
	}

	err = k.UnmarshalWithConf("", out, koanf.UnmarshalConf{Tag: tag, FlatPaths: false})
	if err != nil {
		return configErr(err)
	}

	return nil
}

type Validator interface {
	Validate() error
}

func LoadAndCheckConfig(tag string, out Validator) error {
	err := LoadConfigTo(tag, out)
	if err != nil {
		return err
	}

	if err = out.Validate(); err != nil {
		return configErr(err)
	}

	return nil
}
