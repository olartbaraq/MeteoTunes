package api

import (
	"github.com/spf13/viper"
)

type Config struct {
	AUTH0_DOMAIN     string `mapstructure:"AUTH0_DOMAIN"`
	AUTH0_AUDIENCE   string `mapstructure:"AUTH0_AUDIENCE"`
	PORT             int32  `mapstructure:"PORT"`
	OPEN_WEATHER_KEY string `mapstructure:"OPEN_WEATHER_KEY"`
	LIME_WIRE_KEY    string `mapstructure:"LIME_WIRE_KEY"`
}

func LoadConfig(path string) (config *Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
