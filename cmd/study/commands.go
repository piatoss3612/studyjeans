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

var myStudyInfoCmd = discordgo.ApplicationCommand{
	Name:        "내-정보",
	Description: "내 스터디 등록 정보를 확인합니다.",
}

var registerCmd = discordgo.ApplicationCommand{
	Name:        "발표자-등록",
	Description: "발표자로 등록합니다.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "이름",
			Description: "발표자의 이름을 입력해주세요.",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
		{
			Name:        "주제",
			Description: "발표 주제를 입력해주세요.",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
	},
}
