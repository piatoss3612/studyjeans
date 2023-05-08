package main

import "github.com/bwmarrin/discordgo"

var pingCmd = discordgo.ApplicationCommand{
	Name:        "핑",
	Description: "핑퐁",
}

func (b *StudyBot) pingCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "퐁!",
		},
	})
}

var profileCmd = discordgo.ApplicationCommand{
	Name:        "프로필",
	Description: "발표 진스의 프로필을 보여줍니다.",
}

func (b *StudyBot) profileCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Image: &discordgo.MessageEmbedImage{
						URL: s.State.User.AvatarURL("256"),
					},
				},
			},
		},
	})
}
