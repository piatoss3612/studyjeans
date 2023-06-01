package bot

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/presentation-helper-bot/internal/bot/command"
	"github.com/piatoss3612/presentation-helper-bot/internal/event"
	"github.com/piatoss3612/presentation-helper-bot/internal/event/msgqueue"
	"go.uber.org/zap"
)

type StudyBot struct {
	sess *discordgo.Session
	pub  msgqueue.Publisher

	commands           []*discordgo.ApplicationCommand
	registeredCommands []*discordgo.ApplicationCommand
	handlers           map[string]command.HandleFunc

	startedAt time.Time

	sugar *zap.SugaredLogger
}

func New(pub msgqueue.Publisher, cmdReg command.Registerer, sess *discordgo.Session, sugar *zap.SugaredLogger) *StudyBot {
	return &StudyBot{
		sess:      sess,
		pub:       pub,
		commands:  cmdReg.Commands(),
		handlers:  cmdReg.Handlers(),
		startedAt: time.Now(),
		sugar:     sugar,
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
	var h command.HandleFunc
	var ok bool

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		h, ok = b.handlers[i.ApplicationCommandData().Name]
	case discordgo.InteractionMessageComponent:
		h, ok = b.handlers[i.MessageComponentData().CustomID]
	case discordgo.InteractionModalSubmit:
		h, ok = b.handlers[i.ModalSubmitData().CustomID]
	default:
		return
	}

	if ok {
		if err := h(s, i); err != nil {
			b.sugar.Errorw("failed to handle command", "error", err)
			b.publishError(&event.ErrorEvent{
				T: "study.error",
				D: fmt.Sprintf("%s: %s", i.ApplicationCommandData().Name, err.Error()),
				C: time.Now(),
			})

			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags: discordgo.MessageFlagsEphemeral,
					Embeds: []*discordgo.MessageEmbed{
						ErrorEmbed(err.Error()),
					},
				},
			})
		}
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

func (b *StudyBot) publishError(evt event.Event) {
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

func ErrorEmbed(msg string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "오류",
		Description: msg,
		Color:       0xff0000,
	}
}
