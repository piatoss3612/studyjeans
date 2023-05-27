package config

import "github.com/spf13/viper"

type StudyConfig struct {
	Discord struct {
		BotToken  string `mapstructure:"bot_token"`
		GuildID   string `mapstructure:"guild_id"`
		ManagerID string `mapstructure:"manager_id"`
	} `mapstructure:"discord"`
	MongoDB struct {
		URI    string `mapstructure:"uri"`
		DBName string `mapstructure:"db_name"`
	} `mapstructure:"mongodb"`
	Redis struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"redis"`
	RabbitMQ struct {
		Addr     string `mapstructure:"addr"`
		Exchange string `mapstructure:"exchange"`
		Kind     string `mapstructure:"kind"`
	} `mapstructure:"rabbitmq"`
}

func NewStudyConfig(filename string) (*StudyConfig, error) {
	viper.SetConfigFile(filename)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg StudyConfig

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
