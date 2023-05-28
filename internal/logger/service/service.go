package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/piatoss3612/presentation-helper-bot/internal/event"
	"github.com/piatoss3612/presentation-helper-bot/internal/logger"
	"google.golang.org/api/sheets/v4"
)

var infoLabelFormat = &sheets.CellFormat{
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

type Service interface {
	RecordRound(ctx context.Context, r logger.Round) error
	RecordEvent(ctx context.Context, e event.Event) error
}

type sheetsService struct {
	s             *sheets.Service
	spreadsheetID string
	eventSheetID  int64
}

func New(ctx context.Context, s *sheets.Service, spreadsheetID string, eventSheetID int64) (Service, error) {
	svc := &sheetsService{
		s:             s,
		spreadsheetID: spreadsheetID,
		eventSheetID:  eventSheetID,
	}

	return svc.setup(ctx)
}

func (svc *sheetsService) setup(ctx context.Context) (Service, error) {
	// get spreadsheet
	resp, err := svc.s.Spreadsheets.Get(svc.spreadsheetID).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	// check status code
	if resp.HTTPStatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code while getting spreadsheet: %d", resp.HTTPStatusCode)
	}

	// check event sheet exists
	var eventSheetExists bool

	for _, sheet := range resp.Sheets {
		if sheet.Properties.SheetId == svc.eventSheetID {
			eventSheetExists = true
			break
		}
	}

	if !eventSheetExists {
		// create event sheet
		addSheetReq := &sheets.AddSheetRequest{
			Properties: &sheets.SheetProperties{
				Title:     "이벤트 로그",
				SheetId:   svc.eventSheetID,
				SheetType: "GRID",
			},
		}

		appendCellsReq := &sheets.AppendCellsRequest{
			SheetId: svc.eventSheetID,
			Fields:  "*",
			Rows: []*sheets.RowData{
				{
					Values: []*sheets.CellData{
						{
							UserEnteredFormat: infoLabelFormat,
							UserEnteredValue: &sheets.ExtendedValue{
								StringValue: func() *string {
									s := "이벤트 이름"
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

		resp, err := svc.s.Spreadsheets.BatchUpdate(svc.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
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
			return nil, err
		}

		// check status code
		if resp.HTTPStatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code while adding sheet: %d", resp.HTTPStatusCode)
		}
	}

	return svc, nil
}

func (svc *sheetsService) RecordRound(ctx context.Context, r logger.Round) error {
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

	rows := svc.rowsFromRoundData(r)

	appendCellsReq := &sheets.AppendCellsRequest{
		SheetId: int64(r.Number),
		Fields:  "*",
		Rows:    rows,
	}

	resp, err := svc.s.Spreadsheets.BatchUpdate(svc.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
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

func (svc *sheetsService) RecordEvent(ctx context.Context, e event.Event) error {
	// record event
	resp, err := svc.s.Spreadsheets.BatchUpdate(svc.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AppendCells: &sheets.AppendCellsRequest{
					SheetId: svc.eventSheetID,
					Fields:  "*",
					Rows: []*sheets.RowData{
						{
							Values: []*sheets.CellData{
								{
									UserEnteredValue: &sheets.ExtendedValue{
										StringValue: func() *string {
											s := e.Topic()
											return &s
										}(),
									},
								},
								{
									UserEnteredValue: &sheets.ExtendedValue{
										StringValue: func() *string {
											s := e.Description()
											return &s
										}(),
									},
								},
								{
									UserEnteredValue: &sheets.ExtendedValue{
										StringValue: func() *string {
											s := time.Unix(e.Timestamp(), 0).Format(time.RFC3339)
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

func (svc *sheetsService) rowsFromRoundData(r logger.Round) []*sheets.RowData {
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
