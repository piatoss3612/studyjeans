package profile

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/my-study-bot/internal/bot/command"
	"github.com/piatoss3612/my-study-bot/internal/utils"
	"go.uber.org/zap"
)

type profileCommand struct {
	startedAt time.Time

	sugar *zap.SugaredLogger
}

func NewProfileCommand(sugar *zap.SugaredLogger) command.Command {
	return &profileCommand{
		startedAt: time.Now(),
		sugar:     sugar,
	}
}

func (p *profileCommand) Register(reg command.Registerer) {
	reg.RegisterCommand(cmd, p.showBotProfile)
}

// show the profile of the bot
func (p *profileCommand) showBotProfile(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	u := s.State.User
	createdAt, _ := utils.FormatSnowflakeToTime(u.ID)
	rebootedAt := utils.FormatRebootDate(p.startedAt)
	uptime := utils.FormatUptime(p.startedAt)

	// show the profile
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: u.Mention(),
			Embeds: []*discordgo.MessageEmbed{
				ProfileEmbed(u, "발표 진스의 프로필", createdAt, rebootedAt, uptime),
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}
