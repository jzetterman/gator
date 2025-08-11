package config

import (
	"encoding/json"
	"os"
)

const configFilename = ".gatorconfig.json"

type Config struct {
	Dburl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func (cfg Config) ReadConfig() (Config, error) {
	configFile, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}
	config, err := os.ReadFile(configFile)
	if err != nil {
		return Config{}, err
	}

	var configStruct Config
	err = json.Unmarshal(config, &configStruct)
	if err != nil {
		return Config{}, err
	}
	return configStruct, nil
}

func (cfg Config) SetUser(username string) error {
	cfg.CurrentUserName = username
	return write(cfg)
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return homeDir + "/" + configFilename, nil
}

func write(cfg Config) error {
	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}
	configJson, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, configJson, 0644)
}
