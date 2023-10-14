package command

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

var (
	c1 = &StubCommand{
		cmd: &discordgo.ApplicationCommand{
			Name: "test",
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
			Name: "test3",
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

	errCreateCommand = errors.New("error")
	errDeleteCommand = errors.New("error")
)

type StubCommandManager struct{}

func (m *StubCommandManager) CommandCreate(guildID string, cmd *discordgo.ApplicationCommand) error {
	return nil
}

func (m *StubCommandManager) CommandDelete(guildID string, cmdID string) error {
	return nil
}

type ErrorCommandManager struct{}

func (m *ErrorCommandManager) CommandCreate(guildID string, cmd *discordgo.ApplicationCommand) error {
	return errCreateCommand
}

func (m *ErrorCommandManager) CommandDelete(guildID string, cmdID string) error {
	return errDeleteCommand
}

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
	return NewCommandRegistry(&StubCommandManager{})
}

func newErrorCommandRegistry() *CommandRegistry {
	return NewCommandRegistry(&ErrorCommandManager{})
}

func TestNewCommandRegistry(t *testing.T) {
	r := newCommandRegistry()
	if r.m == nil {
		t.Errorf("expected m to not be nil")
	}
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

func TestCommandRegistry_CreateCommands(t *testing.T) {
	r := newCommandRegistry()

	r.RegisterCommands(c1, c2)

	err := r.CreateCommands()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCommandRegistry_Handle(t *testing.T) {
	r := newCommandRegistry()

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
		err := r.Handle(nil, i)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	}

	err := r.Handle(nil, &discordgo.InteractionCreate{
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

	err = r.Handle(nil, &discordgo.InteractionCreate{
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

func TestCommandRegistry_DeleteCommands(t *testing.T) {
	r := newCommandRegistry()

	r.RegisterCommands(c1, c2)

	err := r.CreateCommands()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	err = r.DeleteCommands()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCommandRegistry_CreateAndDeleteCommandsError(t *testing.T) {
	r := newErrorCommandRegistry()

	r.RegisterCommands(c1, c2)

	err := r.CreateCommands()
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if err != errCreateCommand {
		t.Errorf("expected err to be %v, got %v", errCreateCommand, err)
	}

	err = r.DeleteCommands()
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if err != errDeleteCommand {
		t.Errorf("expected err to be %v, got %v", errDeleteCommand, err)
	}
}
