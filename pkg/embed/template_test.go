package embed

import (
	"errors"
	"testing"
)

func TestErrorEmbed(t *testing.T) {
	err := errors.New("test error")
	embed := ErrorEmbed(err)

	if embed.Title != "Error" {
		t.Errorf("expected embed title to be 'Error', got '%s'", embed.Title)
	}
	if embed.Description != "test error" {
		t.Errorf("expected embed description to be 'test error', got '%s'", embed.Description)
	}
	if embed.Color != 0xff0000 {
		t.Errorf("expected embed color to be 0xff0000, got '%d'", embed.Color)
	}
}
