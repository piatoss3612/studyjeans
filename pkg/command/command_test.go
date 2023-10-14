package command

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"

	_ "github.com/joho/godotenv/autoload"
)

var (
	c1 = &StubCommand{
		cmd: &discordgo.ApplicationCommand{
			Name:        "test",
			Description: "test",
		},
		f: func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
			return nil
		},
		h: map[string]CommandHandleFunc{
			"test2": func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
				return nil
			},
		},
	}

	c2 = &StubCommand{
		cmd: &discordgo.ApplicationCommand{
			Name:        "test3",
			Description: "test3",
		},
		f: func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
			return nil
		},
		h: map[string]CommandHandleFunc{
			"test4": func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
				return nil
			},
		},
	}
)

type StubCommand struct {
	cmd *discordgo.ApplicationCommand
	f   CommandHandleFunc
	h   map[string]CommandHandleFunc
}

func (c *StubCommand) Command() *discordgo.ApplicationCommand {
	return c.cmd
}

func (c *StubCommand) HandleFunc() CommandHandleFunc {
	return c.f
}

func (c *StubCommand) InteractionHandleFuncs() map[string]CommandHandleFunc {
	return c.h
}

func newCommandRegistry() *CommandRegistry {
	return NewCommandRegistry()
}

func TestNewCommandRegistry(t *testing.T) {
	r := newCommandRegistry()
	if len(r.cmds) != 0 {
		t.Errorf("expected cmds to be empty, got %v", r.cmds)
	}
	if len(r.handlers) != 0 {
		t.Errorf("expected handlers to be empty, got %v", r.handlers)
	}
}

func TestCommandRegistry_RegisterCommand(t *testing.T) {
	r := newCommandRegistry()

	r.RegisterCommand(c1)

	if len(r.cmds) != 1 {
		t.Errorf("expected cmds to have 1 element, got %v", r.cmds)
	}

	if r.cmds[0].Name != "test" {
		t.Errorf("expected cmds[0].Name to be \"test\", got %v", r.cmds[0].Name)
	}

	if len(r.handlers) != 2 {
		t.Errorf("expected handlers to have 2 elements, got %v", r.handlers)
	}
}

func TestCommandRegistry_RegisterCommands(t *testing.T) {
	r := newCommandRegistry()

	r.RegisterCommands(c1, c2)

	if len(r.cmds) != 2 {
		t.Errorf("expected cmds to have 2 elements, got %v", r.cmds)
	}

	if r.cmds[0].Name != "test" {
		t.Errorf("expected cmds[0].Name to be \"test\", got %v", r.cmds[0].Name)
	}

	if r.cmds[1].Name != "test3" {
		t.Errorf("expected cmds[1].Name to be \"test3\", got %v", r.cmds[1].Name)
	}

	if len(r.handlers) != 4 {
		t.Errorf("expected handlers to have 4 elements, got %v", r.handlers)
	}
}

func TestCommandRegistry_Commands(t *testing.T) {
	r := newCommandRegistry()

	r.RegisterCommands(c1, c2)

	cmds := r.Commands()

	if len(cmds) != 2 {
		t.Errorf("expected cmds to have 2 elements, got %v", cmds)
	}

	if cmds[0].Name != "test" {
		t.Errorf("expected cmds[0].Name to be \"test\", got %v", cmds[0].Name)
	}

	if cmds[1].Name != "test3" {
		t.Errorf("expected cmds[1].Name to be \"test3\", got %v", cmds[1].Name)
	}
}

func TestCommandRegistry_Handlers(t *testing.T) {
	r := newCommandRegistry()

	r.RegisterCommands(c1, c2)

	handlers := r.Handlers()

	if len(handlers) != 4 {
		t.Errorf("expected handlers to have 4 elements, got %v", handlers)
	}
}

func TestNewCommandManager(t *testing.T) {
	s := &discordgo.Session{}
	r := newCommandRegistry()

	m := NewCommandManager(s, r)

	if m.s != s {
		t.Errorf("expected m.s to be s, got %v", m.s)
	}

	if m.r != r {
		t.Errorf("expected m.r to be r, got %v", m.r)
	}
}

func TestCommandManager_CommandRegistry(t *testing.T) {
	s := &discordgo.Session{}
	r := newCommandRegistry()

	m := NewCommandManager(s, r)

	if m.CommandRegistry() != r {
		t.Errorf("expected m.CommandRegistry() to be r, got %v", m.CommandRegistry())
	}
}

func TestCommandManager_SetCommandRegistry(t *testing.T) {
	s := &discordgo.Session{}
	r := newCommandRegistry()

	m := NewCommandManager(s, r)

	r2 := newCommandRegistry()

	m.SetCommandRegistry(r2)

	if m.CommandRegistry() != r2 {
		t.Errorf("expected m.CommandRegistry() to be r2, got %v", m.CommandRegistry())
	}
}

func TestCommandManager_CreateAndDeleteCommands(t *testing.T) {
	s, err := discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
	if err != nil {
		t.Errorf("expected err to be nil, got %v", err)
	}
	r := newCommandRegistry()

	c := &StubCommand{
		cmd: &discordgo.ApplicationCommand{
			Name:        "hello",
			Description: "hello",
		},
		f: func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
			return nil
		},
	}

	r.RegisterCommands(c)

	m := NewCommandManager(s, r)

	err = s.Open()
	if err != nil {
		t.Errorf("expected err to be nil, got %v", err)
	}
	defer s.Close()

	err = m.CreateCommands("")
	if err != nil {
		t.Errorf("expected err to be nil, got %v", err)
	}

	err = m.DeleteCommands("")
	if err == nil {
		t.Errorf("expected err to be not nil, got %v", err)
	}
}

func TestCommandManager_Handle(t *testing.T) {
	r := newCommandRegistry()

	m := NewCommandManager(nil, r)

	cmd := &StubCommand{
		cmd: &discordgo.ApplicationCommand{
			Name: "application_command",
		},
		f: func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
			return nil
		},
		h: map[string]CommandHandleFunc{
			"message_component": func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
				return nil
			},
			"modal_submit": func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
				return nil
			},
		},
	}

	interactions := []*discordgo.InteractionCreate{
		{
			Interaction: &discordgo.Interaction{
				Type: discordgo.InteractionApplicationCommand,
				Data: discordgo.ApplicationCommandInteractionData{
					Name: "application_command",
				},
			},
		},
		{
			Interaction: &discordgo.Interaction{
				Type: discordgo.InteractionMessageComponent,
				Data: discordgo.MessageComponentInteractionData{
					CustomID: "message_component",
				},
			},
		},
		{
			Interaction: &discordgo.Interaction{
				Type: discordgo.InteractionModalSubmit,
				Data: discordgo.ModalSubmitInteractionData{
					CustomID: "modal_submit",
				},
			},
		},
	}

	r.RegisterCommand(cmd)

	for _, i := range interactions {
		err := m.Handle(nil, i)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	}

	err := m.Handle(nil, &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionPing,
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "unknown",
			},
		},
	})
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), fmt.Sprintf("interaction type %v not found", discordgo.InteractionPing)) {
		t.Errorf("expected error to contain %v, got %v", fmt.Sprintf("interaction type %v not found", discordgo.InteractionPing), err)
	}

	err = m.Handle(nil, &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "unknown",
			},
		},
	})
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), fmt.Sprintf("handler for interaction %s not found", "unknown")) {
		t.Errorf("expected error to contain %v, got %v", fmt.Sprintf("handler for interaction %s not found", "unknown"), err)
	}
}
