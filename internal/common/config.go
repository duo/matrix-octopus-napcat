package common

import (
	"os"
	"time"

	"go.mau.fi/zeroconfig"
	"gopkg.in/yaml.v3"
)

const (
	defaultInitTimeout    = 15 * time.Second
	defaultRequestTimeout = 1 * time.Minute
	defaultPingInterval   = 30 * time.Second
)

type Configure struct {
	NapCat struct {
		ManagerPort    uint32        `yaml:"manager_port"`
		Path           string        `yaml:"path"`
		CMD            string        `yaml:"cmd"`
		InitTimeout    time.Duration `yaml:"init_timeout"`
		RequestTimeout time.Duration `yaml:"request_timeout"`
	}

	Octopus struct {
		ID           string        `yaml:"id"`
		Addr         string        `yaml:"addr"`
		Secret       string        `yaml:"secret"`
		PingInterval time.Duration `yaml:"ping_interval"`
	} `yaml:"octopus"`

	Logging zeroconfig.Config `yaml:"logging"`
}

func LoadConfig(path string) (*Configure, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Configure{}
	config.NapCat.InitTimeout = defaultInitTimeout
	config.NapCat.RequestTimeout = defaultRequestTimeout
	config.Octopus.PingInterval = defaultPingInterval
	if err := yaml.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	return config, nil
}
