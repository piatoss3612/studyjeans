package main

import "github.com/bwmarrin/discordgo"

var helpCmd = discordgo.ApplicationCommand{
	Name:        "도움",
	Description: "도움말을 확인합니다.",
}

var profileCmd = discordgo.ApplicationCommand{
	Name:        "프로필",
	Description: "발표 진스의 프로필을 보여줍니다.",
}
