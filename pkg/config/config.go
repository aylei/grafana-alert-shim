package config

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

var conf = Config{
	Addr: "0.0.0.0",
	Port: 9086,
}

type Config struct {
	Addr   string       `yaml:"addr"`
	Port   int          `yaml:"port"`
	Reader ReaderConfig `yaml:"reader"`
	Writer WriterConfig `yaml:"writer"`
}

type ReaderConfig struct {
	// Type, possible values: generic
	Type string `yaml:"type"`

	Generic *GenericReaderConfig `yaml:"generic"`
}

type GenericReaderConfig struct {
	RulerBaseURL   string `yaml:"rulerBaseURL"`
	QuerierBaseURL string `yaml:"querierBaseURL"`
}

type WriterConfig struct {
	// Type, possible values: configmap
	Type string `yaml:"type"`

	ConfigMap *ConfigMapWriterConfig `yaml:"configmap"`
}

type ConfigMapWriterConfig struct {
	Namespace string `yaml:"namespace"`
	Name      string `yaml:"name"`
}

func LoadConfig(fp string) error {
	fd, err := os.Open(fp)
	if err != nil {
		return err
	}
	bs, err := ioutil.ReadAll(fd)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(bs, &conf)
	if err != nil {
		return err
	}
	if conf.Writer.Type == "" {
		return errors.New("writer.type must be specified")
	}
	if conf.Reader.Type == "" {
		return errors.New("reader.type must be specified")
	}
	return nil
}

func GetConfig() *Config {
	return &conf
}
