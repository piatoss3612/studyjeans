package event

import (
	"fmt"
	"time"
)

type StudyEvent struct {
	Title     string    `json:"title"`
	Stage     Stage     `json:"stage"`
	CreatedAt time.Time `json:"created_at"`
}

func (s StudyEvent) Name() string {
	return s.Stage.String()
}

func (s StudyEvent) Description() string {
	return fmt.Sprintf("%s: %s", s.Title, s.Stage.String())
}

func (s StudyEvent) Timestamp() int64 {
	return s.CreatedAt.Unix()
}

type Stage uint8

const (
	StageNone                 Stage = 0
	StageWait                 Stage = 1
	StageRegistrationOpened   Stage = 2
	StageRegistrationClosed   Stage = 3
	StageSubmissionOpened     Stage = 4
	StageSubmissionClosed     Stage = 5
	StagePresentationStarted  Stage = 6
	StagePresentationFinished Stage = 7
	StageReviewOpened         Stage = 8
	StageReviewClosed         Stage = 9
)

func (s Stage) String() string {
	switch s {
	case StageWait:
		return "다음 라운드 대기"
	case StageRegistrationOpened:
		return "발표자 등록"
	case StageRegistrationClosed:
		return "발표자 등록 마감"
	case StageSubmissionOpened:
		return "발표 자료 제출"
	case StageSubmissionClosed:
		return "발표 자료 제출 마감"
	case StagePresentationStarted:
		return "발표"
	case StagePresentationFinished:
		return "발표 종료"
	case StageReviewOpened:
		return "피드백"
	case StageReviewClosed:
		return "피드백 마감"
	default:
		return "초기화"
	}
}

var _ = Event(&StudyEvent{})
