package submit

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/presentation-helper-bot/internal/bot/command"
	"github.com/piatoss3612/presentation-helper-bot/internal/study"
	"github.com/piatoss3612/presentation-helper-bot/internal/study/service"
	"github.com/piatoss3612/presentation-helper-bot/internal/utils"
	"go.uber.org/zap"
)

type submitCommand struct {
	svc service.Service

	sugar *zap.SugaredLogger
}

func NewSubmitCommand(svc service.Service, sugar *zap.SugaredLogger) command.Command {
	return &submitCommand{
		svc:   svc,
		sugar: sugar,
	}
}

func (sc *submitCommand) Register(reg command.Registerer) {
	reg.RegisterCommand(cmd, sc.submitContent)
}

// submit content for presentation
func (sc *submitCommand) submitContent(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	user := utils.GetGuildUserFromInteraction(i)
	if user == nil {
		return study.ErrUserNotFound
	}

	// get content
	var content string

	for _, option := range i.ApplicationCommandData().Options {
		switch option.Name {
		case "링크":
			content = option.StringValue()
		}
	}

	if content == "" {
		return errors.Join(study.ErrRequiredArgs, errors.New("발표 자료 링크는 필수 입력 사항입니다"))
	}

	_, err := url.Parse(content)
	if err != nil {
		return errors.Join(study.ErrInvalidArgs, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// set content
	_, _, err = sc.svc.UpdateRound(ctx, &service.UpdateParams{
		GuildID:    i.GuildID,
		MemberID:   user.ID,
		ContentURL: content,
	},
		service.SubmitMemberContent, service.ValidateToSubmitMemberContent)
	if err != nil {
		return err
	}

	// send response
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: user.Mention(),
			Flags:   discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				submitEmbed(s.State.User, "제출 완료", "발표 자료가 제출되었습니다.", content),
			},
		},
	})
}
