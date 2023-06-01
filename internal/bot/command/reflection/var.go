package reflection

import "github.com/bwmarrin/discordgo"

var cmd = discordgo.ApplicationCommand{
	Name:        "발표회고",
	Description: "발표회고를 작성합니다.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "내용",
			Description: "발표회고 내용을 입력해주세요.",
			Required:    true,
		},
	},
}

func ErrorEmbed(msg string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "오류",
		Description: msg,
		Color:       0xff0000,
	}
}
