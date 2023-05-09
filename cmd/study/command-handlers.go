package main

import (
	"context"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/presentation-helper-bot/internal/utils"
)

func (b *StudyBot) helpCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

func (b *StudyBot) registerCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := b.svc.GetGuildID()

	if i.GuildID != guildID {
		// TODO: error response
		return
	}

	var user *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		user = i.Member.User
	}

	if user == nil {
		// TODO: error response
		return
	}

	var name, subject string

	for _, option := range i.ApplicationCommandData().Options {
		switch option.Name {
		case "이름":
			name = option.StringValue()
		case "주제":
			subject = option.StringValue()
		}
	}

	if name == "" || subject == "" {
		// TODO: error response
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := b.svc.ChangeMemberRegistration(ctx, i.GuildID, user.ID, name, subject, true)
	if err != nil {
		// TODO: error response
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: user.Mention(),
			Flags:   discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				EmbedTemplate(s.State.User, "등록 완료", "발표자 등록이 완료되었습니다."),
			},
		},
	})
	if err != nil {
		// TODO: error response
		return
	}
}

func (b *StudyBot) submitContentCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := b.svc.GetGuildID()

	if i.GuildID != guildID {
		// TODO: error response
		log.Println("guildID not matched")
		return
	}

	var user *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		user = i.Member.User
	}

	if user == nil {
		// TODO: error response
		log.Println("user not found")
		return
	}

	var content string

	for _, option := range i.ApplicationCommandData().Options {
		switch option.Name {
		case "링크":
			content = option.StringValue()
		}
	}

	if content == "" {
		// TODO: error response
		log.Println("content not found")
		return
	}

	// TODO: validate if content is url

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := b.svc.SubmitContent(ctx, i.GuildID, user.ID, content)
	if err != nil {
		// TODO: error response
		log.Println(err)
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: user.Mention(),
			Flags:   discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				EmbedTemplate(s.State.User, "제출 완료", "발표 자료가 제출되었습니다."),
			},
		},
	})
	if err != nil {
		// TODO: error response
		log.Println(err)
		return
	}
}

func (b *StudyBot) sendFeedbackCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := b.svc.GetGuildID()

	if i.GuildID != guildID {
		// TODO: error response
		log.Println("guildID not matched")
		return
	}

	var user *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		user = i.Member.User
	}

	if user == nil {
		// TODO: error response
		log.Println("user not found")
		return
	}

	var presentor *discordgo.User

	for _, option := range i.ApplicationCommandData().Options {
		switch option.Name {
		case "발표자":
			presentor = option.UserValue(s)
		}
	}

	if presentor == nil {
		// TODO: error response
		log.Println("presentor not found")
		return
	}

	if user.ID == presentor.ID {
		// TODO: error response
		log.Println("user and presentor is same")
		return
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "feedback-modal",
			Title:    "피드백 작성",
			Flags:    discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "id",
							Label:       "발표자",
							Style:       discordgo.TextInputShort,
							Placeholder: "발표자의 ID 입니다. 임의로 변경하지 마세요.",
							Value:       presentor.ID,
							Required:    true,
							MaxLength:   20,
							MinLength:   1,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "feedback",
							Label:       "피드백",
							Style:       discordgo.TextInputParagraph,
							Placeholder: "피드백을 입력해주세요.",
							Required:    true,
							MaxLength:   1000,
							MinLength:   10,
						},
					},
				},
			},
		},
	})
	if err != nil {
		// TODO: error response
		log.Println(err)
	}
}
