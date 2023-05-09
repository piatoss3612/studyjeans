package study

import (
	"context"
	"errors"
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
		_ = errorInteractionRespond(s, i, err)
	}
}

func (b *StudyBot) profileCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	u := s.State.User
	createdAt, _ := utils.FormatSnowflakeToTime(u.ID)
	rebootedAt := utils.FormatRebootDate(b.startedAt)
	uptime := utils.FormatUptime(b.startedAt)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: u.Mention(),
			Embeds: []*discordgo.MessageEmbed{
				BotInfoEmbed(u, "발표 진스의 프로필", createdAt, rebootedAt, uptime),
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		_ = errorInteractionRespond(s, i, err)
	}
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

		study, err := b.svc.GetOngoingStudy(ctx, i.GuildID)
		if err != nil {
			return err
		}

		if study == nil {
			return ErrStudyNotFound
		}

		member, ok := study.GetMember(user.ID)
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
		_ = errorInteractionRespond(s, i, err)
	}
}

func (b *StudyBot) registerCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmd := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		var user *discordgo.User

		if i.Member != nil && i.Member.User != nil {
			user = i.Member.User
		}

		if user == nil {
			return ErrUserNotFound
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
			return errors.Join(ErrRequiredArgs, errors.New("이름과 발표 주제는 필수 입력 사항입니다."))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := b.svc.SetMemberRegistered(ctx, i.GuildID, user.ID, name, subject, true)
		if err != nil {
			return err
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: user.Mention(),
				Flags:   discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{
					EmbedTemplate(s.State.User, "등록 완료", "발표자 등록이 완료되었습니다."),
				},
			},
		})
	}

	err := cmd(s, i)
	if err != nil {
		_ = errorInteractionRespond(s, i, err)
	}
}

func (b *StudyBot) unregisterCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

		err := b.svc.SetMemberRegistered(ctx, i.GuildID, user.ID, "", "", false)
		if err != nil {
			return err
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: user.Mention(),
				Flags:   discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{
					EmbedTemplate(s.State.User, "등록 취소 완료", "발표자 등록이 취소되었습니다."),
				},
			},
		})
	}

	err := cmd(s, i)
	if err != nil {
		_ = errorInteractionRespond(s, i, err)
	}
}

func (b *StudyBot) submitContentCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmd := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		var user *discordgo.User

		if i.Member != nil && i.Member.User != nil {
			user = i.Member.User
		}

		if user == nil {
			return ErrUserNotFound
		}

		var content string

		for _, option := range i.ApplicationCommandData().Options {
			switch option.Name {
			case "링크":
				content = option.StringValue()
			}
		}

		if content == "" {
			return errors.Join(ErrRequiredArgs, errors.New("발표 자료 링크는 필수 입력 사항입니다."))
		}

		// TODO: validate if content is url

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		err := b.svc.SubmitContent(ctx, i.GuildID, user.ID, content)
		if err != nil {
			return err
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: user.Mention(),
				Flags:   discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{
					EmbedTemplate(s.State.User, "제출 완료", "발표 자료가 제출되었습니다."),
				},
			},
		})
	}

	err := cmd(s, i)
	if err != nil {
		_ = errorInteractionRespond(s, i, err)
	}
}

func (b *StudyBot) sendFeedbackCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmd := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		var user *discordgo.User

		if i.Member != nil && i.Member.User != nil {
			user = i.Member.User
		}

		if user == nil {
			return ErrUserNotFound
		}

		var presentor *discordgo.User

		for _, option := range i.ApplicationCommandData().Options {
			switch option.Name {
			case "발표자":
				presentor = option.UserValue(s)
			}
		}

		if presentor == nil {
			return errors.Join(ErrRequiredArgs, errors.New("리뷰 대상자는 필수 입력 사항입니다."))
		}

		if presentor.Bot {
			return errors.New("봇은 리뷰 대상자로 지정할 수 없습니다.")
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: "feedback-modal",
				Title:    "피드백 작성",
				Flags:    discordgo.MessageFlagsEphemeral,
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    "presentor-id",
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
						Components: []discordgo.MessageComponent{feedbackTextInput},
					},
				},
			},
		})
	}

	err := cmd(s, i)
	if err != nil {
		_ = errorInteractionRespond(s, i, err)
	}
}
