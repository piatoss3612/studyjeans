package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	cerrors "github.com/piatoss3612/my-study-bot/internal/errors"
	"github.com/piatoss3612/my-study-bot/internal/pubsub"
	"google.golang.org/api/sheets/v4"
)

var (
	defaultErrorSheetID int64 = 400
	errorLabelFormat          = &sheets.CellFormat{
		TextFormat: &sheets.TextFormat{
			Bold: true,
			ForegroundColor: &sheets.Color{
				Red:   1.0,
				Green: 1.0,
				Blue:  1.0,
			},
		},
		BackgroundColor: &sheets.Color{
			Red: 0.8,
		},
		HorizontalAlignment: "CENTER",
	}
)

type handler struct {
	s             *sheets.Service
	spreadsheetID string
	errorSheetID  int64
}

type HandlerOptsFunc func(*handler)

func WithDefaultErrorSheetID() HandlerOptsFunc {
	return func(h *handler) {
		h.errorSheetID = defaultErrorSheetID
	}
}

func WithErrorSheetID(id int64) HandlerOptsFunc {
	return func(h *handler) {
		h.errorSheetID = id
	}
}

func NewHandler(ctx context.Context, s *sheets.Service, spreadsheetID string, opts ...HandlerOptsFunc) (pubsub.Handler, error) {
	h := &handler{
		s:             s,
		spreadsheetID: spreadsheetID,
		errorSheetID:  defaultErrorSheetID,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h.setup(ctx)
}

// setup error sheet
func (h *handler) setup(ctx context.Context) (pubsub.Handler, error) {
	// get spreadsheet
	resp, err := h.s.Spreadsheets.Get(h.spreadsheetID).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	// check status code
	if resp.HTTPStatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code while getting spreadsheet: %d", resp.HTTPStatusCode)
	}

	// check event sheet exists
	var progressSheetExists bool

	for _, sheet := range resp.Sheets {
		if sheet.Properties.SheetId == h.errorSheetID {
			progressSheetExists = true
		}
	}

	if !progressSheetExists {
		// create progress sheet
		if err := h.createErrorSheet(ctx); err != nil {
			return nil, err
		}
	}

	return h, nil
}

func (h *handler) Handle(ctx context.Context, body []byte) error {
	evt := cerrors.Event{}

	if err := json.Unmarshal(body, &evt); err != nil {
		return err
	}

	switch evt.Topic {
	case cerrors.EventTopicError:
		return h.RecordError(ctx, evt)
	default:
		return errors.Join(cerrors.ErrUnknownEventTopic, fmt.Errorf("unknown event topic: %s", evt.Topic))
	}
}

// create error sheet
func (h *handler) createErrorSheet(ctx context.Context) error {
	addSheetReq := &sheets.AddSheetRequest{
		Properties: &sheets.SheetProperties{
			Title:     "에러 로그",
			SheetId:   h.errorSheetID,
			SheetType: "GRID",
		},
	}

	appendCellsReq := &sheets.AppendCellsRequest{
		SheetId: h.errorSheetID,
		Fields:  "*",
		Rows: []*sheets.RowData{
			{
				Values: []*sheets.CellData{
					{
						UserEnteredFormat: errorLabelFormat,
						UserEnteredValue: &sheets.ExtendedValue{
							StringValue: func() *string {
								s := "오류"
								return &s
							}(),
						},
					},
					{
						UserEnteredFormat: errorLabelFormat,
						UserEnteredValue: &sheets.ExtendedValue{
							StringValue: func() *string {
								s := "설명"
								return &s
							}(),
						},
					},
					{
						UserEnteredFormat: errorLabelFormat,
						UserEnteredValue: &sheets.ExtendedValue{
							StringValue: func() *string {
								s := "시간"
								return &s
							}(),
						},
					},
				},
			},
		},
	}

	resp, err := h.s.Spreadsheets.BatchUpdate(h.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AddSheet: addSheetReq,
			},
			{
				AppendCells: appendCellsReq,
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return err
	}

	// check status code
	if resp.HTTPStatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code while adding sheet: %d", resp.HTTPStatusCode)
	}

	return nil
}

func (h *handler) RecordError(ctx context.Context, evt cerrors.Event) error {
	resp, err := h.s.Spreadsheets.BatchUpdate(h.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AppendCells: &sheets.AppendCellsRequest{
					SheetId: h.errorSheetID,
					Fields:  "*",
					Rows: []*sheets.RowData{
						{
							Values: []*sheets.CellData{
								{
									UserEnteredValue: &sheets.ExtendedValue{
										StringValue: func() *string {
											s := evt.Topic.String()
											return &s
										}(),
									},
								},
								{
									UserEnteredValue: &sheets.ExtendedValue{
										StringValue: func() *string {
											s := evt.Description
											return &s
										}(),
									},
								},
								{
									UserEnteredValue: &sheets.ExtendedValue{
										StringValue: func() *string {
											s := time.Unix(evt.Timestamp, 0).Format(time.RFC3339)
											return &s
										}(),
									},
								},
							},
						},
					},
				},
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return err
	}

	// check status code
	if resp.HTTPStatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.HTTPStatusCode)
	}

	return nil
}
