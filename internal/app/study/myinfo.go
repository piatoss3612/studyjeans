package study

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/presentation-helper-bot/internal/models/study"
)

var myStudyInfoCmd = discordgo.ApplicationCommand{
	Name:        "내-정보",
	Description: "내 스터디 회차 등록 정보를 확인합니다.",
}

func (b *StudyBot) addMyStudyInfoCmd() {
	b.hdr.AddCommand(myStudyInfoCmd, b.myStudyInfoCmdHandler)
}

func (b *StudyBot) myStudyInfoCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmd := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		var user *discordgo.User

		if i.Member != nil && i.Member.User != nil {
			user = i.Member.User
		}

		if user == nil {
			return ErrUserNotFound
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		round, err := b.svc.GetOngoingRound(ctx, i.GuildID)
		if err != nil {
			return err
		}

		if round == nil {
			return ErrRoundNotFound
		}

		member, ok := round.GetMember(user.ID)
		if !ok {
			return ErrMemberNotFound
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: user.Mention(),
				Flags:   discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{
					MyStudyInfoEmbed(user, member),
				},
			},
		})
	}

	err := cmd(s, i)
	if err != nil {
		b.sugar.Errorw(err.Error(), "event", "my-study-info")
		_ = errorInteractionRespond(s, i, err)
	}
}

func MyStudyInfoEmbed(u *discordgo.User, m study.Member) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "나의 스터디 등록 정보",
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: u.AvatarURL(""),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "이름",
				Value: func() string {
					if m.Name == "" {
						return "```미등록```"
					}
					return fmt.Sprintf("```%s```", m.Name)
				}(),
				Inline: true,
			},
			{
				Name: "발표자 등록",
				Value: func() string {
					if m.Registered {
						return "```O```"
					}
					return "```X```"
				}(),
				Inline: true,
			},
			{
				Name: "발표 완료",
				Value: func() string {
					if m.Attended {
						return "```O```"
					}
					return "```X```"
				}(),
				Inline: true,
			},
			{
				Name: "발표주제",
				Value: func() string {
					if m.Subject == "" {
						return "```미등록```"
					}
					return fmt.Sprintf("```%s```", m.Subject)
				}(),
			},
			{
				Name: "발표자료",
				Value: func() string {
					if m.ContentURL == "" {
						return "```미등록```"
					}
					return fmt.Sprintf("```%s```", m.ContentURL)
				}(),
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Color:     16777215,
	}
}
