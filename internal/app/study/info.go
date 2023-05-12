package study

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/presentation-helper-bot/internal/models/study"
)

var (
	myStudyInfoCmd = discordgo.ApplicationCommand{
		Name:        "내-정보",
		Description: "내 스터디 회차 등록 정보를 확인합니다.",
	}
	studyRoundInfoCmd = discordgo.ApplicationCommand{
		Name:        "라운드-정보",
		Description: "현재 진행중인 스터디 라운드 정보를 확인합니다.",
	}
	speakerInfoSelectMenu = discordgo.SelectMenu{
		CustomID:    "speaker-info",
		Placeholder: "발표자 등록 정보 검색 🔍",
		MenuType:    discordgo.UserSelectMenu,
	}
)

func (b *StudyBot) addStudyInfoCmd() {
	b.hdr.AddCommand(myStudyInfoCmd, b.myStudyInfoCmdHandler)
	b.hdr.AddCommand(studyRoundInfoCmd, b.studyRoundInfoCmdHandler)
	b.chdr.AddHandleFunc(speakerInfoSelectMenu.CustomID, b.speakerInfoSelectMenuHandler)
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
					SpeakerInfoEmbed(user, member),
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

func (b *StudyBot) studyRoundInfoCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmd := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		var user *discordgo.User

		// command should be invoked only in guild
		if i.Member != nil && i.Member.User != nil {
			user = i.Member.User
		}

		if user == nil {
			return ErrUserNotFound
		}

		// TODO: load cached round info or fetch from db if not exists

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		round, err := b.svc.GetOngoingRound(ctx, i.GuildID)
		if err != nil {
			return err
		}

		embed := studyRoundInfoEmbed(s.State.User, round)

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:  discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{embed},
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							speakerInfoSelectMenu,
						},
					},
				},
			},
		})
	}

	err := cmd(s, i)
	if err != nil {
		b.sugar.Errorw(err.Error(), "event", "study-round-info")
		_ = errorInteractionRespond(s, i, err)
	}
}

func (b *StudyBot) speakerInfoSelectMenuHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO: implement
	cmd := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		var user *discordgo.User

		// command should be invoked only in guild
		if i.Member != nil && i.Member.User != nil {
			user = i.Member.User
		}

		if user == nil {
			return ErrUserNotFound
		}

		// get data
		data := i.MessageComponentData().Values
		if len(data) == 0 {
			return errors.Join(ErrRequiredArgs, errors.New("옵션을 찾을 수 없습니다"))
		}

		selectedUserID := data[0]

		// TODO: load cached round info or fetch from db if not exists

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		round, err := b.svc.GetOngoingRound(ctx, i.GuildID)
		if err != nil {
			return err
		}

		if round == nil {
			return ErrRoundNotFound
		}

		var embed *discordgo.MessageEmbed

		member, ok := round.GetMember(selectedUserID)
		if !ok {
			embed = ErrorEmbed("발표자 정보를 찾을 수 없습니다")
		} else {
			embed = SpeakerInfoEmbed(user, member)
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Flags:  discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{embed},
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							speakerInfoSelectMenu,
						},
					},
				},
			},
		})
	}

	err := cmd(s, i)
	if err != nil {
		b.sugar.Errorw(err.Error(), "event", "study-round-info")
		_ = errorInteractionRespond(s, i, err)
	}
}

func studyRoundInfoEmbed(u *discordgo.User, r *study.Round) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    u.Username,
			IconURL: u.AvatarURL(""),
		},
		Title:     "현재 진행중인 스터디 라운드 정보",
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: u.AvatarURL("")},
		Fields: []*discordgo.MessageEmbedField{

			{
				Name:   "번호",
				Value:  fmt.Sprintf("```%d```", r.Number),
				Inline: true,
			},
			{
				Name:   "제목",
				Value:  fmt.Sprintf("```%s```", r.Title),
				Inline: true,
			},
			{
				Name:   "진행 단계",
				Value:  fmt.Sprintf("```%s```", r.Stage.String()),
				Inline: true,
			},
			{
				Name: "발표 결과 자료",
				Value: fmt.Sprintf("```%s```", func() string {
					if r.ContentURL == "" {
						return "미등록"
					}
					return r.ContentURL
				}()),
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func SpeakerInfoEmbed(u *discordgo.User, m study.Member) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "발표자 등록 정보",
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
				Name: "발표 주제",
				Value: func() string {
					if m.Subject == "" {
						return "```미등록```"
					}
					return fmt.Sprintf("```%s```", m.Subject)
				}(),
			},
			{
				Name: "발표 자료",
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
