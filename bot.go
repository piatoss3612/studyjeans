package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/presentation-helper-bot/study"
)

type StudyBot struct {
	sess *discordgo.Session
	svc  *study.Service
}

func NewStudyBot(sess *discordgo.Session, svc *study.Service) *StudyBot {
	return &StudyBot{sess: sess, svc: svc}
}

func (b *StudyBot) Setup() *StudyBot {
	b.sess.Identify.Intents = discordgo.IntentGuildMembers | discordgo.IntentGuildMessages | discordgo.IntentGuilds | discordgo.IntentDirectMessages
	return b
}

func (b *StudyBot) Run() (<-chan bool, error) {
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

func (b *StudyBot) Close() error {
	return b.sess.Close()
}
