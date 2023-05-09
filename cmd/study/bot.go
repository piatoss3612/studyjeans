package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	handler "github.com/piatoss3612/presentation-helper-bot/internal/handler/study"
	"github.com/piatoss3612/presentation-helper-bot/internal/service/study"
	"go.uber.org/zap"
)

type StudyBot struct {
	sess *discordgo.Session
	hdr  handler.Handler
	chdr handler.ComponentHandler
	svc  study.Service

	guildID   string
	startedAt time.Time

	sugar *zap.SugaredLogger
}

func NewStudyBot(sess *discordgo.Session, svc study.Service, sugar *zap.SugaredLogger) *StudyBot {
	return &StudyBot{
		sess:      sess,
		hdr:       handler.New(),
		chdr:      handler.NewComponent(),
		svc:       svc,
		startedAt: time.Now(),
		sugar:     sugar,
	}
}

func (b *StudyBot) Setup() *StudyBot {
	b.sess.Identify.Intents = discordgo.IntentGuildMembers | discordgo.IntentGuildMessages | discordgo.IntentGuilds | discordgo.IntentDirectMessages

	b.sess.AddHandler(b.ready)
	b.sess.AddHandler(b.handleApplicationCommand)

	b.hdr.AddCommand(adminCmd, b.adminHandler)
	b.hdr.AddCommand(helpCmd, b.helpCmdHandler)
	b.hdr.AddCommand(profileCmd, b.profileCmdHandler)
	b.hdr.AddCommand(myStudyInfoCmd, b.myStudyInfoCmdHandler)
	b.hdr.AddCommand(registerCmd, b.registerCmdHandler)
	b.hdr.AddCommand(unregisterCmd, b.unregisterCmdHandler)
	b.hdr.AddCommand(submitContentCmd, b.submitContentCmdHandler)
	b.hdr.AddCommand(sendFeedbackCmd, b.sendFeedbackCmdHandler)

	b.chdr.AddHandleFunc(helpSelectMenu.CustomID, b.helpSelectMenuHandler)
	b.chdr.AddHandleFunc("feedback-modal", b.feedbackSubmitHandler)

	return b
}

func (b *StudyBot) Run() (<-chan bool, error) {
	if err := b.sess.Open(); err != nil {
		return nil, err
	}

	if err := b.hdr.RegisterApplicationCommands(b.sess); err != nil {
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
	if err := b.hdr.RemoveApplicationCommands(b.sess); err != nil {
		return err
	}

	return b.sess.Close()
}

func (b *StudyBot) ready(s *discordgo.Session, _ *discordgo.Ready) {
	_ = s.UpdateGameStatus(0, "초기화")
}

func (b *StudyBot) handleApplicationCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var h handler.HandleFunc
	var ok bool

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		h, ok = b.hdr.GetHandleFunc(i.ApplicationCommandData().Name)
	case discordgo.InteractionMessageComponent:
		h, ok = b.chdr.GetHandleFunc(i.MessageComponentData().CustomID)
	case discordgo.InteractionModalSubmit:
		h, ok = b.chdr.GetHandleFunc(i.ModalSubmitData().CustomID)
	default:
		return
	}

	if ok {
		h(s, i)
	}
}
