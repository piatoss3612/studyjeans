package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/presentation-helper-bot/internal/utils"
)

func (b *StudyBot) helpHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:  discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{HelpIntroEmbed(s.State.User)},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						helpSelectMenu,
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Emoji: discordgo.ComponentEmoji{
								Name: "🔥",
							},
							Label: "큰 결심 하기",
							Style: discordgo.LinkButton,
							URL:   "https://github.com/piatoss3612",
						},
					},
				},
			},
		},
	})

	if err != nil {
		log.Println(err)
	}
}

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
				BotInfoEmbed(u, "발표 진스의 프로필", createdAt, rebootedAt, uptime),
			},
		},
	})
}

func (b *StudyBot) myStudyInfoCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := b.svc.GetGuildID()

	if i.GuildID != guildID {
		// TODO: error response
	}

	var user *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		user = i.Member.User
	}

	if user == nil {
		// TODO: error response
	}

	member, ok := b.svc.GetMember(user.ID)
	if !ok {
		// TODO: error response
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: user.Mention(),
			Flags:   discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				MyStudyInfoEmbed(user, member),
			},
		},
	})
	if err != nil {
		// TODO: error response
	}
}
