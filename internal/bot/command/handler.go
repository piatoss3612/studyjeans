package command

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/my-study-bot/internal/event"
	"github.com/piatoss3612/my-study-bot/internal/msgqueue"
	"go.uber.org/zap"
)

type Handler interface {
	Handle(name string, s *discordgo.Session, i *discordgo.InteractionCreate)
}

type handler struct {
	funcs map[string]HandleFunc
	pub   msgqueue.Publisher

	sugar *zap.SugaredLogger
}

func NewHandler(funcs map[string]HandleFunc, pub msgqueue.Publisher, sugar *zap.SugaredLogger) Handler {
	return &handler{
		funcs: funcs,
		pub:   pub,
		sugar: sugar,
	}
}

func (h *handler) Handle(name string, s *discordgo.Session, i *discordgo.InteractionCreate) {
	fn, ok := h.funcs[name]
	if !ok {
		return
	}

	// TODO: improve error handling
	if err := fn(s, i); err != nil {
		h.sugar.Errorw("failed to handle command", "error", err)
		h.publishError(&event.ErrorEvent{
			T: "study.error",
			D: fmt.Sprintf("%s: %s", name, err.Error()),
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

func (h *handler) publishError(evt event.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		err := h.pub.Publish(ctx, evt.Topic(), evt)
		if err != nil {
			h.sugar.Errorw(err.Error(), "event", "publish event", "topic", evt.Topic(), "try", i+1)
			continue
		}
		return
	}
}
