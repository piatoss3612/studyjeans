package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	handler "github.com/piatoss3612/presentation-helper-bot/internal/handler/study"
	"github.com/piatoss3612/presentation-helper-bot/internal/service/study"
)

type StudyBot struct {
	sess *discordgo.Session
	hdr  handler.Handler
	svc  study.Service
}

func NewStudyBot(sess *discordgo.Session, svc study.Service) *StudyBot {
	return &StudyBot{sess: sess, hdr: handler.NewHandler(), svc: svc}
}

func (b *StudyBot) Setup() *StudyBot {
	b.sess.Identify.Intents = discordgo.IntentGuildMembers | discordgo.IntentGuildMessages | discordgo.IntentGuilds | discordgo.IntentDirectMessages

	b.sess.AddHandler(b.ready)
	b.sess.AddHandler(b.handleApplicationCommand)

	b.hdr.AddCommand(pingCmd, b.pingCmdHandler)
	b.hdr.AddCommand(profileCmd, b.profileCmdHandler)

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
	_ = s.UpdateGameStatus(0, b.svc.GetCurrentStudyStage().String())
}

func (b *StudyBot) handleApplicationCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if h, ok := b.hdr.GetHandleFunc(i.ApplicationCommandData().Name); ok {
		h(s, i)
	}
}
