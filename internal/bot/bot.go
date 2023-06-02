package bot

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/my-study-bot/internal/bot/command"
	"github.com/piatoss3612/my-study-bot/internal/event"
	"github.com/piatoss3612/my-study-bot/internal/msgqueue"
	"go.uber.org/zap"
)

type Bot interface {
	Run() (<-chan bool, error)
	RegisterCommands(reg command.Registerer) error
	RemoveCommands() error
	Close() error
}

type bot struct {
	sess               *discordgo.Session
	pub                msgqueue.Publisher
	registeredCommands []*discordgo.ApplicationCommand
	handlers           map[string]command.HandleFunc

	sugar *zap.SugaredLogger
}

func New(pub msgqueue.Publisher, sess *discordgo.Session, sugar *zap.SugaredLogger) Bot {
	b := &bot{
		sess:     sess,
		pub:      pub,
		handlers: map[string]command.HandleFunc{},
		sugar:    sugar,
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

func (b *bot) RegisterCommands(reg command.Registerer) error {
	cmds := reg.Commands()

	registeredCmds := make([]*discordgo.ApplicationCommand, 0, len(cmds))

	for _, cmd := range cmds {
		registered, err := b.sess.ApplicationCommandCreate(b.sess.State.User.ID, "", cmd)
		if err != nil {
			return err
		}

		registeredCmds = append(registeredCmds, registered)
	}

	b.registeredCommands = registeredCmds
	b.handlers = reg.Handlers()

	return nil
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
	var cmdName string

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		cmdName = i.ApplicationCommandData().Name
	case discordgo.InteractionMessageComponent:
		cmdName = i.MessageComponentData().CustomID
	case discordgo.InteractionModalSubmit:
		cmdName = i.ModalSubmitData().CustomID
	default:
		return
	}

	h, ok := b.handlers[cmdName]
	if ok {
		if err := h(s, i); err != nil {
			b.sugar.Errorw("failed to handle command", "error", err)
			b.publishError(&event.ErrorEvent{
				T: "study.error",
				D: fmt.Sprintf("%s: %s", cmdName, err.Error()),
				C: time.Now(),
			})

			embed := &discordgo.MessageEmbed{
				Title:       "오류",
				Description: err.Error(),
				Color:       0xff0000,
			}

			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:  discordgo.MessageFlagsEphemeral,
					Embeds: []*discordgo.MessageEmbed{embed},
				},
			})
		}
	}
}

func (b *bot) publishError(evt event.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		err := b.pub.Publish(ctx, evt.Topic(), evt)
		if err != nil {
			b.sugar.Errorw(err.Error(), "event", "publish event", "topic", evt.Topic(), "try", i+1)
			continue
		}
		return
	}
}
