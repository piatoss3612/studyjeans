package bot

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/my-study-bot/internal/bot/command"
)

type Bot interface {
	Run() (<-chan bool, error)
	RegisterCommands(cmds []*discordgo.ApplicationCommand) error
	RegisterHandler(h command.Handler)
	RemoveCommands() error
	Close() error
}

type bot struct {
	sess               *discordgo.Session
	registeredCommands []*discordgo.ApplicationCommand
	handler            command.Handler
}

func New(sess *discordgo.Session) Bot {
	b := &bot{
		sess: sess,
	}
	return b.setup()
}

func (b *bot) setup() Bot {
	b.sess.Identify.Intents = discordgo.IntentGuildMembers | discordgo.IntentGuildMessages |
		discordgo.IntentGuilds | discordgo.IntentDirectMessages

	b.sess.AddHandler(b.ready)
	b.sess.AddHandler(b.handleApplicationCommand)

	return b
}

func (b *bot) Run() (<-chan bool, error) {
	if err := b.sess.Open(); err != nil {
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

func (b *bot) RegisterCommands(cmds []*discordgo.ApplicationCommand) error {
	registeredCmds := make([]*discordgo.ApplicationCommand, 0, len(cmds))

	for _, cmd := range cmds {
		registered, err := b.sess.ApplicationCommandCreate(b.sess.State.User.ID, "", cmd)
		if err != nil {
			return err
		}

		registeredCmds = append(registeredCmds, registered)
	}

	b.registeredCommands = registeredCmds
	return nil
}

func (b *bot) RegisterHandler(h command.Handler) {
	b.handler = h
}

func (b *bot) RemoveCommands() error {
	appID := b.sess.State.User.ID

	for _, cmd := range b.registeredCommands {
		if err := b.sess.ApplicationCommandDelete(appID, "", cmd.ID); err != nil {
			return err
		}
	}

	return nil
}

func (b *bot) Close() error {
	return b.sess.Close()
}

func (b *bot) ready(s *discordgo.Session, _ *discordgo.Ready) {
	_ = s.UpdateGameStatus(0, "초기화")
}

func (b *bot) handleApplicationCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var name string

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		name = i.ApplicationCommandData().Name
	case discordgo.InteractionMessageComponent:
		name = i.MessageComponentData().CustomID
	case discordgo.InteractionModalSubmit:
		name = i.ModalSubmitData().CustomID
	default:
		return
	}

	b.handler.Handle(name, s, i)
}
