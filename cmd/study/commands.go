package main

import "github.com/bwmarrin/discordgo"

var profileCmd = discordgo.ApplicationCommand{
	Name:        "프로필",
	Description: "발표 진스의 프로필을 보여줍니다.",
}
