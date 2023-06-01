package profile

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/presentation-helper-bot/internal/bot/command"
	"github.com/piatoss3612/presentation-helper-bot/internal/event"
	"github.com/piatoss3612/presentation-helper-bot/internal/event/msgqueue"
	"github.com/piatoss3612/presentation-helper-bot/internal/utils"
	"go.uber.org/zap"
)

type profileCommand struct {
	pub msgqueue.Publisher

	startedAt time.Time

	sugar *zap.SugaredLogger
}

func NewProfileCommand(pub msgqueue.Publisher, sugar *zap.SugaredLogger) command.Command {
	return &profileCommand{
		pub:       pub,
		startedAt: time.Now(),
		sugar:     sugar,
	}
}

func (p *profileCommand) Register(reg command.Registerer) {
	reg.RegisterCommand(cmd, p.profileCmdHandler)
}

func (p *profileCommand) profileCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	u := s.State.User
	createdAt, _ := utils.FormatSnowflakeToTime(u.ID)
	rebootedAt := utils.FormatRebootDate(p.startedAt)
	uptime := utils.FormatUptime(p.startedAt)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: u.Mention(),
			Embeds: []*discordgo.MessageEmbed{
				ProfileEmbed(u, "발표 진스의 프로필", createdAt, rebootedAt, uptime),
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		go func() {
			evt := &event.ErrorEvent{
				T: "study.error",
				D: fmt.Sprintf("%s: %s", i.ApplicationCommandData().Name, err.Error()),
				C: time.Now(),
			}

			go p.publishEvent(evt)
		}()
		p.sugar.Errorw(err.Error(), "event", i.ApplicationCommandData().Name)
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:  discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{ErrorEmbed(err.Error())},
			},
		})
	}
}

func (p *profileCommand) publishEvent(evt event.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		err := p.pub.Publish(ctx, evt.Topic(), evt)
		if err != nil {
			p.sugar.Errorw(err.Error(), "event", "publish event", "topic", evt.Topic(), "try", i+1)
			continue
		}
		return
	}
}
