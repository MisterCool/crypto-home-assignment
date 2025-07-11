package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ChainConfig struct {
	Chain     string `yaml:"chain"`
	RPCUrl    string `yaml:"rpc_url"`
	APIKey    string `yaml:"api_key"`
	StartFrom int64  `yaml:"start_from"`
	BatchSize int64  `yaml:"batch_size"`
}

type AppConfig struct {
	Chains []ChainConfig `yaml:"chains"`
}

func LoadConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
