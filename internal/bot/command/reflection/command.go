package reflection

import (
	"context"
	"errors"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/presentation-helper-bot/internal/bot/command"
	"github.com/piatoss3612/presentation-helper-bot/internal/study"
	"github.com/piatoss3612/presentation-helper-bot/internal/study/service"
	"github.com/piatoss3612/presentation-helper-bot/internal/utils"
	"go.uber.org/zap"
)

type reflectionCommand struct {
	svc service.Service

	sugar *zap.SugaredLogger
}

func NewReflectionCommand(svc service.Service, sugar *zap.SugaredLogger) command.Command {
	return &reflectionCommand{
		svc:   svc,
		sugar: sugar,
	}
}

func (rc *reflectionCommand) Register(reg command.Registerer) {
	reg.RegisterCommand(cmd, rc.sendReflection)
}

func (rc *reflectionCommand) sendReflection(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// user should be in guild
	user := utils.GetGuildUserFromInteraction(i)
	if user == nil {
		return study.ErrUserNotFound
	}

	content := i.ApplicationCommandData().Options[0].StringValue()

	// content should not be empty
	if content == "" {
		return errors.Join(study.ErrRequiredArgs, errors.New("회고 내용은 필수입니다"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// set sent reflection
	gs, _, err := rc.svc.UpdateRound(ctx, &service.UpdateParams{
		GuildID:  i.GuildID,
		MemberID: user.ID,
	},
		service.SetSentReflection, service.ValidateToSetSendReflection)
	if err != nil {
		return err
	}

	if gs.ReflectionChannelID == "" {
		return study.ErrChannelNotFound
	}

	embed := reflectionEmbed(user, content)

	// send reflection
	_, err = s.ChannelMessageSendEmbed(gs.ReflectionChannelID, embed)
	if err != nil {
		return err
	}

	// send response
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "회고가 성공적으로 전송되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
