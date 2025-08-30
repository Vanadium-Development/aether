package config

import (
	"os"
	"reflect"

	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
)

const configPath = "config.toml"

type NodeConfig struct {
	Data struct {
		TempDirectory   string `toml:"temp_directory"`
		ScenesDirectory string `toml:"scenes_directory"`
	} `toml:"Data"`
	Node struct {
		Name string `toml:"node_name"`
		Port uint16 `toml:"port"`
	} `toml:"Node"`
}

func validateConfig(cfg any) {
	cfgValue := reflect.ValueOf(cfg)
	cfgType := reflect.TypeOf(cfg)
	var field reflect.Value
	for i := 0; i < cfgValue.NumField(); i++ {
		field = cfgValue.Field(i)
		if (field.Kind() == reflect.String && field.String() == "") || (field.Kind() == reflect.Uint16 && field.Uint() == 0) {
			logrus.Fatal("Config does not set \"" + cfgType.Field(i).Tag.Get("toml") + "\"")
		}
	}
}

func ParseNodeConfig() NodeConfig {
	var cfg NodeConfig
	_, err := toml.DecodeFile(configPath, &cfg)

	if err != nil {
		logrus.Fatal("Error parsing config file \"" + configPath + "\":" + err.Error())
	}

	validateConfig(cfg.Node)
	validateConfig(cfg.Data)

	return cfg
}

func ensureFolder(path string) {
	// TODO Handle other errors
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		logrus.Debugf("Creating directory: %s", path)
		os.MkdirAll(path, 0755)
	}
}

func (cfg *NodeConfig) EnsureFolders() {
	ensureFolder(cfg.Data.TempDirectory)
	ensureFolder(cfg.Data.ScenesDirectory)
}
