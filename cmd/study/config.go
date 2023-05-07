package main

import "github.com/spf13/viper"

type Config struct {
	Environment     string `mapstructure:"environment"`
	BotToken        string `mapstructure:"bot_token"`
	MongoURI        string `mapstructure:"mongo_uri"`
	DBName          string `mapstructure:"db_name"`
	GuildID         string `mapstructure:"guild_id"`
	ManagerID       string `mapstructure:"manager_id"`
	NoticeChannelID string `mapstructure:"notice_channel_id"`
}

func NewConfig(filename string) (*Config, error) {
	viper.SetConfigFile(filename)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config

	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
