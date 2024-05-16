package main

import (
	"github.com/samber/mo"
	yaml "gopkg.in/yaml.v3"
	"io"
	"os"
)

type Keyword struct {
	Engine  string `yaml:"engine"`
	Prepend string `yaml:"prepend"`
}

type KeywordEither struct{ mo.Either[string, *Keyword] }

type EngineConfig struct {
	Fallback string            `yaml:"fallback"`
	Russian  string            `yaml:"russian"`
	Japanese string            `yaml:"japanese"`
	URLS     map[string]string `yaml:"urls"`
}

type Config struct {
	Port  int  `yaml:"port"`
	Debug bool `yaml:"debug"`

	Engines  EngineConfig              `yaml:"engines"`
	Keywords map[string]*KeywordEither `yaml:"keywords"`
}

func (ke *KeywordEither) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var err error
	var s string
	if err = unmarshal(&s); err == nil {
		*ke = KeywordEither{mo.Left[string, *Keyword](s)}
		return nil
	}

	var k Keyword
	if err = unmarshal(&k); err == nil {
		*ke = KeywordEither{mo.Right[string, *Keyword](&k)}
		return nil
	}

	return err
}

func ReadConfig(fileName string) (*Config, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
