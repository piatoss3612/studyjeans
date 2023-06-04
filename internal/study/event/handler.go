package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/piatoss3612/my-study-bot/internal/pubsub"
	"github.com/piatoss3612/my-study-bot/internal/study"
	"google.golang.org/api/sheets/v4"
)

var (
	defaultProgressSheetID int64 = 1024
	infoLabelFormat              = &sheets.CellFormat{
		TextFormat: &sheets.TextFormat{
			Bold: true,
			ForegroundColor: &sheets.Color{
				Red:   1.0,
				Green: 1.0,
				Blue:  1.0,
			},
		},
		BackgroundColor: &sheets.Color{
			Blue: 0.8,
		},
		HorizontalAlignment: "CENTER",
	}
)

type handler struct {
	s               *sheets.Service
	spreadsheetID   string
	progressSheetID int64
}

type HandlerOptsFunc func(*handler)

func WithDefaultProgressSheetID() HandlerOptsFunc {
	return func(h *handler) {
		h.progressSheetID = defaultProgressSheetID
	}
}

func WithProgressSheetID(id int64) HandlerOptsFunc {
	return func(h *handler) {
		h.progressSheetID = id
	}
}

func New(ctx context.Context, s *sheets.Service, spreadSheetID string, opts ...HandlerOptsFunc) (pubsub.Handler, error) {
	h := &handler{
		s:               s,
		spreadsheetID:   spreadSheetID,
		progressSheetID: defaultProgressSheetID,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h.setup(ctx)
}

// setup progress sheet
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
		if sheet.Properties.SheetId == h.progressSheetID {
			progressSheetExists = true
		}
	}

	if !progressSheetExists {
		// create progress sheet
		if err := h.createProgressSheet(ctx); err != nil {
			return nil, err
		}
	}

	return h, nil
}

func (h *handler) Handle(ctx context.Context, body []byte) error {
	evt := study.Event{}

	if err := json.Unmarshal(body, &evt); err != nil {
		return err
	}

	switch evt.Topic {
	case study.EventTopicStudyRoundCreated, study.EventTopicStudyRoundProgress:
		return h.recordProgress(ctx, evt)
	case study.EventTopicStudyRoundFinished:
		var r study.Round

		if err := json.Unmarshal(evt.Data, &r); err != nil {
			return err
		}

		return h.recordRound(ctx, r)
	default:
		return errors.Join(study.ErrUnknownEventTopic, fmt.Errorf("unknown event topic: %s", evt.Topic))
	}
}

// record round data to spreadsheet
func (h *handler) recordRound(ctx context.Context, r study.Round) error {
	addSheetReq := &sheets.AddSheetRequest{
		Properties: &sheets.SheetProperties{
			Title:     fmt.Sprintf("%d 라운드: %s", r.Number, r.Title),
			SheetId:   int64(r.Number),
			SheetType: "GRID",
			TabColor: &sheets.Color{
				Blue: 1.0,
			},
		},
	}

	rows := rowsFromRoundData(r)

	appendCellsReq := &sheets.AppendCellsRequest{
		SheetId: int64(r.Number),
		Fields:  "*",
		Rows:    rows,
	}

	resp, err := h.s.Spreadsheets.BatchUpdate(h.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AddSheet: addSheetReq, // create sheet
			},
			{
				AppendCells: appendCellsReq, // then add rows
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

func (h *handler) recordProgress(ctx context.Context, evt study.Event) error {
	resp, err := h.s.Spreadsheets.BatchUpdate(h.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AppendCells: &sheets.AppendCellsRequest{
					SheetId: h.progressSheetID,
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

func (h *handler) createProgressSheet(ctx context.Context) error {
	addSheetReq := &sheets.AddSheetRequest{
		Properties: &sheets.SheetProperties{
			Title:     "진행 로그",
			SheetId:   h.progressSheetID,
			SheetType: "GRID",
		},
	}

	appendCellsReq := &sheets.AppendCellsRequest{
		SheetId: h.progressSheetID,
		Fields:  "*",
		Rows: []*sheets.RowData{
			{
				Values: []*sheets.CellData{
					{
						UserEnteredFormat: infoLabelFormat,
						UserEnteredValue: &sheets.ExtendedValue{
							StringValue: func() *string {
								s := "진행 상태"
								return &s
							}(),
						},
					},
					{
						UserEnteredFormat: infoLabelFormat,
						UserEnteredValue: &sheets.ExtendedValue{
							StringValue: func() *string {
								s := "설명"
								return &s
							}(),
						},
					},
					{
						UserEnteredFormat: infoLabelFormat,
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

func rowsFromRoundData(r study.Round) []*sheets.RowData {
	rows := []*sheets.RowData{
		{
			Values: []*sheets.CellData{
				{
					UserEnteredFormat: infoLabelFormat,
					UserEnteredValue: &sheets.ExtendedValue{
						StringValue: func() *string {
							s := "제목"
							return &s
						}(),
					},
				},
				{
					UserEnteredValue: &sheets.ExtendedValue{
						StringValue: func() *string {
							s := r.Title
							return &s
						}(),
					},
				},
			},
		},
		{
			Values: []*sheets.CellData{
				{
					UserEnteredFormat: infoLabelFormat,
					UserEnteredValue: &sheets.ExtendedValue{
						StringValue: func() *string {
							s := "진행 단계"
							return &s
						}(),
					},
				},
				{
					UserEnteredValue: &sheets.ExtendedValue{
						StringValue: func() *string {
							s := r.Stage.String()
							return &s
						}(),
					},
				},
			},
		},
		{
			Values: []*sheets.CellData{
				{
					UserEnteredFormat: infoLabelFormat,
					UserEnteredValue: &sheets.ExtendedValue{
						StringValue: func() *string {
							s := "녹화 영상"
							return &s
						}(),
					},
				},
				{
					UserEnteredValue: &sheets.ExtendedValue{
						FormulaValue: func() *string {
							s := fmt.Sprintf(`=HYPERLINK("%s")`, r.ContentURL)
							return &s
						}(),
					},
				},
			},
		},
		{
			Values: []*sheets.CellData{
				{
					UserEnteredFormat: infoLabelFormat,
					UserEnteredValue: &sheets.ExtendedValue{
						StringValue: func() *string {
							s := "생성"
							return &s
						}(),
					},
				},
				{
					UserEnteredValue: &sheets.ExtendedValue{
						StringValue: func() *string {
							s := r.CreatedAt.Format(time.RFC3339)
							return &s
						}(),
					},
				},
			},
		},
		{
			Values: []*sheets.CellData{
				{
					UserEnteredFormat: infoLabelFormat,
					UserEnteredValue: &sheets.ExtendedValue{
						StringValue: func() *string {
							s := "최종 수정"
							return &s
						}(),
					},
				},
				{
					UserEnteredValue: &sheets.ExtendedValue{
						StringValue: func() *string {
							s := r.UpdatedAt.Format(time.RFC3339)
							return &s
						}(),
					},
				},
			},
		},
		{}, // empty row
		{
			Values: []*sheets.CellData{
				{
					UserEnteredFormat: infoLabelFormat,
					UserEnteredValue: &sheets.ExtendedValue{
						StringValue: func() *string {
							s := "ID"
							return &s
						}(),
					},
				},
				{
					UserEnteredFormat: infoLabelFormat,
					UserEnteredValue: &sheets.ExtendedValue{
						StringValue: func() *string {
							s := "이름"
							return &s
						}(),
					},
				},
				{
					UserEnteredFormat: infoLabelFormat,
					UserEnteredValue: &sheets.ExtendedValue{
						StringValue: func() *string {
							s := "발표 주제"
							return &s
						}(),
					},
				},
				{
					UserEnteredFormat: infoLabelFormat,
					UserEnteredValue: &sheets.ExtendedValue{
						StringValue: func() *string {
							s := "발표 자료"
							return &s
						}(),
					},
				},
				{
					UserEnteredFormat: infoLabelFormat,
					UserEnteredValue: &sheets.ExtendedValue{
						StringValue: func() *string {
							s := "발표 참여"
							return &s
						}(),
					},
				},
			},
		},
	}

	for id, m := range r.Members {
		row := &sheets.RowData{
			Values: []*sheets.CellData{
				{
					UserEnteredValue: &sheets.ExtendedValue{
						StringValue: func() *string {
							s := id
							return &s
						}(),
					},
				},
				{
					UserEnteredValue: &sheets.ExtendedValue{
						StringValue: func() *string {
							s := m.Name
							return &s
						}(),
					},
				},
				{
					UserEnteredValue: &sheets.ExtendedValue{
						StringValue: func() *string {
							s := m.Subject
							return &s
						}(),
					},
				},
				{
					UserEnteredValue: &sheets.ExtendedValue{
						FormulaValue: func() *string {
							s := fmt.Sprintf(`=HYPERLINK("%s")`, m.ContentURL)
							return &s
						}(),
					},
				},
				{
					UserEnteredValue: &sheets.ExtendedValue{
						BoolValue: func() *bool {
							b := m.Attended
							return &b
						}(),
					},
				},
			},
		}

		rows = append(rows, row)
	}

	return rows
}
