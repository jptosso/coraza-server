package config

import (
	"fmt"
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Config struct {
	LogLevel string  `yaml:"log_level"`
	Agents   []Agent `yaml:"agents"`
}

type Agent struct {
	Protocol       string   `yaml:"protocol"`
	Bind           string   `yaml:"bind"`
	TransactionTtl int      `yaml:"transaction_ttl"`
	Include        []string `yaml:"include"`
	Workers        int      `yaml:"workers"`
	TxActiveLimit  int      `yaml:"transactions_active_limit"`
}

func ReadFile(file string) (Config, error) {
	f, err := os.ReadFile(file)
	var cfg Config
	if err != nil {
		return Config{}, err
	}
	err = yaml.Unmarshal(f, &cfg)
	if err != nil {
		return cfg, err
	}
	if err := ValidateConfig(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func ValidateConfig(cfg Config) error {
	if cfg.LogLevel != "" {
		lvl, err := logrus.ParseLevel(cfg.LogLevel)
		if err != nil {
			return err
		}
		logrus.SetLevel(lvl)
	}
	for _, a := range cfg.Agents {
		//Check protocol
		//check bind
		if a.TransactionTtl < 0 {
			return fmt.Errorf("transaction ttl cannot be less than 0")
		}
		if a.TxActiveLimit < 0 {
			return fmt.Errorf("active transaction limit cannot be less than 0")
		}
		if a.Workers < 0 {
			return fmt.Errorf("workers cannot be less than 0")
		}
		if a.Workers > runtime.NumCPU() {
			logrus.Warnf("You are using more workers available CPU cores, try to lower it to %d", runtime.NumCPU()-1)
		}
	}
	return nil
}
