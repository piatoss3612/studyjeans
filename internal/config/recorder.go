package config

import "github.com/spf13/viper"

type RecorderConfig struct {
	RabbitMQ struct {
		Addr     string   `mapstructure:"addr"`
		Exchange string   `mapstructure:"exchange"`
		Kind     string   `mapstructure:"kind"`
		Queue    string   `mapstructure:"queue"`
		Topics   []string `mapstructure:"topics"`
	} `mapstructure:"rabbitmq"`
}

func NewRecorderConfig(filename string) (*RecorderConfig, error) {
	viper.SetConfigFile(filename)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg RecorderConfig

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
