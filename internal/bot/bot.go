package bot

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/my-study-bot/internal/bot/command"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var totalRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "command_requests_total",
		Help: "Total number of requests.",
	},
	[]string{"command"},
)

var totalErrors = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "command_errors_total",
		Help: "Total number of errors.",
	},
	[]string{"command"},
)

var duration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "command_response_time_seconds",
		Help: "Duration of command response in seconds.",
	},
	[]string{"command"},
)

var metricServerPort = "8080"

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

	srv *http.Server

	sugar *zap.SugaredLogger
}

func New(sess *discordgo.Session, sugar *zap.SugaredLogger) Bot {
	b := &bot{
		sess:  sess,
		sugar: sugar,
	}
	return b.setup()
}

func (b *bot) setup() Bot {
	b.sess.Identify.Intents = discordgo.IntentGuildMembers | discordgo.IntentGuildMessages |
		discordgo.IntentGuilds | discordgo.IntentDirectMessages

	b.sess.AddHandler(b.ready)
	b.sess.AddHandler(b.handleApplicationCommand)

	metrics := prometheus.NewRegistry()
	metrics.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	metrics.MustRegister(totalRequests)
	metrics.MustRegister(totalErrors)
	metrics.MustRegister(duration)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(metrics, promhttp.HandlerOpts{}))
	mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	b.srv = &http.Server{
		Addr:    fmt.Sprintf(":%s", metricServerPort),
		Handler: mux,
	}

	return b
}

func (b *bot) Run() (<-chan bool, error) {
	if err := b.sess.Open(); err != nil {
		return nil, err
	}

	go func() {
		if err := b.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			b.sugar.Fatal("Failed to start metric server", "error", err)
		}
	}()

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
	err := b.sess.Close()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = b.srv.Shutdown(ctx)
	if err != nil {
		return err
	}

	return nil
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

	timer := prometheus.NewTimer(duration.WithLabelValues(name))

	err := b.handler.Handle(name, s, i)

	totalRequests.WithLabelValues(name).Inc()

	if err != nil {
		totalErrors.WithLabelValues(name).Inc()
		b.errorResponse(s, i, err)
		b.sugar.Errorw("command error", "command", name, "error", err.Error(), "duration", timer.ObserveDuration().String())
		return
	}
	b.sugar.Infow("command handled", "command", name, "duration", timer.ObserveDuration().String())
}

func (b *bot) errorResponse(s *discordgo.Session, i *discordgo.InteractionCreate, err error) {
	embed := &discordgo.MessageEmbed{
		Title:       "오류",
		Description: err.Error(),
		Color:       0xff0000,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:  discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}
