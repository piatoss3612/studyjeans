package bot

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/presentation-helper-bot/internal/bot/command"
	"github.com/piatoss3612/presentation-helper-bot/internal/bot/component"
	"go.uber.org/zap"
)

type StudyBot struct {
	sess *discordgo.Session

	commands           []*discordgo.ApplicationCommand
	registeredCommands []*discordgo.ApplicationCommand
	commandHandlers    map[string]command.HandleFunc
	componentHandlers  map[string]component.HandleFunc

	startedAt time.Time

	sugar *zap.SugaredLogger
}

func New(cmdReg command.Registerer, cptReg component.Registerer, sess *discordgo.Session, sugar *zap.SugaredLogger) *StudyBot {
	return &StudyBot{
		sess:              sess,
		commands:          cmdReg.Commands(),
		commandHandlers:   cmdReg.Handlers(),
		componentHandlers: cptReg.Handlers(),
		startedAt:         time.Now(),
		sugar:             sugar,
	}
}

func (b *StudyBot) Setup() *StudyBot {
	b.sess.Identify.Intents = discordgo.IntentGuildMembers | discordgo.IntentGuildMessages |
		discordgo.IntentGuilds | discordgo.IntentDirectMessages

	b.sess.AddHandler(b.ready)
	b.sess.AddHandler(b.handleApplicationCommand)

	return b
}

func (b *StudyBot) Run() (<-chan bool, error) {
	if err := b.sess.Open(); err != nil {
		return nil, err
	}

	if err := b.registerCommands(); err != nil {
		return nil, err
	}

	stop := make(chan bool)
	shutdown := make(chan os.Signal, 1)

	signal.Notify(shutdown,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	go func() {
		defer func() {
			close(shutdown)
			close(stop)
		}()
		<-shutdown
	}()

	return stop, nil
}

func (b *StudyBot) Close() error {
	if err := b.removeRegisteredCommands(); err != nil {
		return err
	}

	return b.sess.Close()
}

func (b *StudyBot) ready(s *discordgo.Session, _ *discordgo.Ready) {
	_ = s.UpdateGameStatus(0, "초기화")
}

func (b *StudyBot) handleApplicationCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var h func(*discordgo.Session, *discordgo.InteractionCreate)
	var ok bool

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		h, ok = b.commandHandlers[i.ApplicationCommandData().Name]
	case discordgo.InteractionMessageComponent:
		h, ok = b.componentHandlers[i.MessageComponentData().CustomID]
	case discordgo.InteractionModalSubmit:
		h, ok = b.componentHandlers[i.ModalSubmitData().CustomID]
	default:
		return
	}

	if ok {
		h(s, i)
	}
}

func (b *StudyBot) registerCommands() error {
	b.registeredCommands = make([]*discordgo.ApplicationCommand, len(b.commands))

	for _, cmd := range b.commands {
		registered, err := b.sess.ApplicationCommandCreate(b.sess.State.User.ID, "", cmd)
		if err != nil {
			return err
		}

		b.registeredCommands = append(b.registeredCommands, registered)
	}

	return nil
}

func (b *StudyBot) removeRegisteredCommands() error {
	for _, cmd := range b.registeredCommands {
		if err := b.sess.ApplicationCommandDelete(b.sess.State.User.ID, "", cmd.ID); err != nil {
			return err
		}
	}

	return nil
}

func errorInteractionRespond(s *discordgo.Session, i *discordgo.InteractionCreate, err error) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:  discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{ErrorEmbed(err.Error())},
		},
	})
}

func EmbedTemplate(u *discordgo.User, title, description string, url ...string) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    u.Username,
			IconURL: u.AvatarURL(""),
		},
		Title:       title,
		Description: description,
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       16777215,
	}

	if len(url) > 0 {
		embed.URL = url[0]
	}

	return embed
}

func ErrorEmbed(msg string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "오류",
		Description: msg,
		Color:       0xff0000,
	}
}
