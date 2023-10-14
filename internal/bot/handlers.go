package bot

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/studyjeans/pkg/command"
	"github.com/piatoss3612/studyjeans/pkg/embed"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

func NewApplicationCommandHandler(m *command.CommandManager, l *zap.Logger) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		totalRequests.WithLabelValues("all").Inc()

		var name string

		switch i.Type {
		case discordgo.InteractionApplicationCommand, discordgo.InteractionApplicationCommandAutocomplete:
			name = i.ApplicationCommandData().Name
		case discordgo.InteractionMessageComponent:
			name = i.MessageComponentData().CustomID
		case discordgo.InteractionModalSubmit:
			name = i.ModalSubmitData().CustomID
		default:
			totalErrors.WithLabelValues(name).Inc()
			l.Error("unknown interaction type", zap.Uint8("type", uint8(i.Type)))
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags: discordgo.MessageFlagsEphemeral,
					Embeds: []*discordgo.MessageEmbed{
						embed.ErrorEmbed(errors.New("unknown interaction type")),
					},
				},
			})
		}

		timer := prometheus.NewTimer(duration.WithLabelValues(name))

		err := m.Handle(name, s, i)
		if err != nil {
			totalErrors.WithLabelValues(name).Inc()
			l.Error("failed to handle interaction", zap.Error(err))
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags: discordgo.MessageFlagsEphemeral,
					Embeds: []*discordgo.MessageEmbed{
						embed.ErrorEmbed(err),
					},
				},
			})
			return
		}

		totalSuccess.WithLabelValues(name).Inc()
		l.Info("handled interaction",
			zap.String("interaction", name),
			zap.String("user", i.User.String()),
			zap.String("guild", i.GuildID),
			zap.String("channel", i.ChannelID),
			zap.Duration("duration", timer.ObserveDuration()),
		)
	}
}
