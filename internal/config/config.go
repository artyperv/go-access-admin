package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type AppSettings struct {
	Debug                 bool `yaml:"debug"`
	SyncHtpasswd          bool `yaml:"sync_htpasswd"`
	CleanAccessesInterval int  `yaml:"clean_accesses_interval"`
}

type HtpasswdPath struct {
	Name        string      `yaml:"name"`
	Path        string      `yaml:"path"`
	URLTemplate string      `yaml:"url_template"`
	Admins      []AdminUser `yaml:"admins"`
}

type AdminUser struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Config struct {
	AppSettings   AppSettings    `yaml:"app"`
	HtpasswdPaths []HtpasswdPath `yaml:"htpasswd_paths"`
	Admins        []AdminUser    `yaml:"admins"`
}

func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
