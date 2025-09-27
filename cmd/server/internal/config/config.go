package config

import "github.com/BurntSushi/toml"


type ProxyServer struct {
	Name  string `toml:"name"`
	Token string `toml:"token"`
	Port  int `toml:"port"`
}

type Config struct {
	ClientPort int `toml:"client_port"`
	ControlPort int `toml:"control_port"`
	Servers []ProxyServer `toml:"servers"`
}

func LoadConfig(file string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(file, &config); err != nil {
		return nil, err
	}

	return &config, nil
}