package config

import "github.com/spf13/viper"

type LoggerConfig struct {
	RabbitMQ struct {
		Addr     string `mapstructure:"addr"`
		Exchange string `mapstructure:"exchange"`
		Kind     string `mapstructure:"kind"`
		Queue    string `mapstructure:"queue"`
	} `mapstructure:"rabbitmq"`
}

func NewLoggerConfig(filename string) (*LoggerConfig, error) {
	viper.SetConfigFile(filename)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg LoggerConfig

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
