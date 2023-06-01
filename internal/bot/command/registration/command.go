package registration

import (
	"context"
	"errors"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/presentation-helper-bot/internal/bot/command"
	"github.com/piatoss3612/presentation-helper-bot/internal/study"
	"github.com/piatoss3612/presentation-helper-bot/internal/study/service"
	"go.uber.org/zap"
)

type registrationCmd struct {
	svc service.Service

	sugar *zap.SugaredLogger
}

func NewRegistrationCommand(svc service.Service, sugar *zap.SugaredLogger) command.Command {
	return &registrationCmd{
		svc:   svc,
		sugar: sugar,
	}
}

func (rc *registrationCmd) Register(reg command.Registerer) {
	reg.RegisterCommand(registerCmd, rc.registerCmdHandler)
	reg.RegisterCommand(unregisterCmd, rc.unregisterCmdHandler)
}

func (rc *registrationCmd) registerCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var user *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		user = i.Member.User
	}

	if user == nil {
		return study.ErrUserNotFound
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
		return errors.Join(study.ErrRequiredArgs, errors.New("이름과 발표 주제는 필수 입력 사항입니다"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, err := rc.svc.UpdateRound(ctx, &service.UpdateParams{
		GuildID:    i.GuildID,
		MemberID:   user.ID,
		MemberName: name,
		Subject:    subject,
	},
		service.RegisterMember, service.ValidateToRegister)
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

func (rc *registrationCmd) unregisterCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var user *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		user = i.Member.User
	}

	if user == nil {
		return study.ErrUserNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, err := rc.svc.UpdateRound(ctx, &service.UpdateParams{
		GuildID:  i.GuildID,
		MemberID: user.ID,
	},
		service.UnregisterSpeaker, service.ValidateToUnregister)
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
