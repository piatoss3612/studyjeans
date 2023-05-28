package bot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/presentation-helper-bot/internal/event"
	"github.com/piatoss3612/presentation-helper-bot/internal/study"
	"github.com/piatoss3612/presentation-helper-bot/internal/study/service"
)

var (
	registerCmd = discordgo.ApplicationCommand{
		Name:        "발표자-등록",
		Description: "발표자 정보를 등록합니다.",
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

	unregisterCmd = discordgo.ApplicationCommand{
		Name:        "발표자-등록-취소",
		Description: "발표자 등록을 취소합니다.",
	}
)

func (b *StudyBot) addRegistrationCmd() {
	b.cmd.AddCommand(registerCmd, b.registerCmdHandler)
	b.cmd.AddCommand(unregisterCmd, b.unregisterCmdHandler)
}

func (b *StudyBot) registerCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	fn := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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

		_, _, err := b.svc.UpdateRound(ctx, &service.UpdateParams{
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

	err := fn(s, i)
	if err != nil {
		go func() {
			evt := &event.ErrorEvent{
				T: "study.error",
				D: fmt.Sprintf("%s: %s", i.ApplicationCommandData().Name, err.Error()),
				C: time.Now(),
			}

			go b.publishEvent(evt)
		}()
		b.sugar.Errorw(err.Error(), "event", i.ApplicationCommandData().Name)
		_ = errorInteractionRespond(s, i, err)
	}
}

func (b *StudyBot) unregisterCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	fn := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		var user *discordgo.User

		if i.Member != nil && i.Member.User != nil {
			user = i.Member.User
		}

		if user == nil {
			return study.ErrUserNotFound
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, _, err := b.svc.UpdateRound(ctx, &service.UpdateParams{
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

	err := fn(s, i)
	if err != nil {
		go func() {
			evt := &event.ErrorEvent{
				T: "study.error",
				D: fmt.Sprintf("%s: %s", i.ApplicationCommandData().Name, err.Error()),
				C: time.Now(),
			}

			go b.publishEvent(evt)
		}()
		b.sugar.Errorw(err.Error(), "event", i.ApplicationCommandData().Name)
		_ = errorInteractionRespond(s, i, err)
	}
}
