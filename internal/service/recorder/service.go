package recorder

import (
	"context"

	"github.com/piatoss3612/presentation-helper-bot/internal/models/event"
	models "github.com/piatoss3612/presentation-helper-bot/internal/models/recorder"
	"google.golang.org/api/sheets/v4"
)

type Service interface {
	RecordRound(ctx context.Context, r models.Round) error
	RecordEvent(ctx context.Context, e event.Event) error
}

type serviceImpl struct {
	s             *sheets.Service
	spreadsheetID string
	eventSheetID  string
}

func NewService(s *sheets.Service, spreadsheetID, eventSheetID string) Service {
	return &serviceImpl{
		s:             s,
		spreadsheetID: spreadsheetID,
		eventSheetID:  eventSheetID,
	}
}

func (svc *serviceImpl) RecordRound(ctx context.Context, r models.Round) error {
	return nil
}

func (svc *serviceImpl) RecordEvent(ctx context.Context, e event.Event) error {
	return nil
}
