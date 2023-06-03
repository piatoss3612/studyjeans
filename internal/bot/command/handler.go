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

	start := time.Now()

	if err := fn(s, i); err != nil {
		h.sugar.Errorw("failed to handle command", "command", name, "error", err.Error(), "duration", time.Since(start).String())

		go h.publishError(&event.ErrorEvent{
			T: "study.error",
			D: fmt.Sprintf("%s: %s", name, err.Error()),
			C: time.Now(),
		})

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
		return
	}

	h.sugar.Infow("command handled", "command", name, "duration", time.Since(start).String())
}

func (h *handler) publishError(evt event.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cnt := 0

	for {
		select {
		case <-ctx.Done():
			h.sugar.Errorw("failed to publish error event", "error", ctx.Err().Error(), "topic", evt.Topic(), "description", evt.Description(), "retry", cnt)
			return
		default:
			err := h.pub.Publish(ctx, evt.Topic(), evt)
			if err != nil {
				h.sugar.Errorw("failed to publish error event", "error", err.Error(), "topic", evt.Topic(), "description", evt.Description(), "retry", cnt)
				time.Sleep(500 * time.Millisecond)
				cnt++
				continue
			}
			h.sugar.Infow("error event published", "topic", evt.Topic(), "retry", cnt)
			return
		}
	}
}
