package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/presentation-helper-bot/internal/utils"
)

func (b *StudyBot) profileCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	u := s.State.User
	createdAt, _ := utils.FormatSnowflakeToTime(u.ID)
	rebootedAt := utils.FormatRebootDate(b.startedAt)
	uptime := utils.FormatUptime(b.startedAt)

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: u.Mention(),
			Embeds: []*discordgo.MessageEmbed{
				InfoEmbed(u, "발표 진스의 프로필", createdAt, rebootedAt, uptime),
			},
		},
	})
}
