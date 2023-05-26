package study

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

func (s Stage) IsNone() bool {
	return s == StageNone
}

func (s Stage) IsWait() bool {
	return s == StageWait
}

func (s Stage) CanMoveTo(target Stage) bool {
	switch s {
	case StageWait:
		return target == StageRegistrationOpened
	case StageRegistrationOpened:
		return target == StageRegistrationClosed
	case StageRegistrationClosed:
		return target == StageSubmissionOpened
	case StageSubmissionOpened:
		return target == StageSubmissionClosed
	case StageSubmissionClosed:
		return target == StagePresentationStarted
	case StagePresentationStarted:
		return target == StagePresentationFinished
	case StagePresentationFinished:
		return target == StageReviewOpened
	case StageReviewOpened:
		return target == StageReviewClosed
	default:
		return false
	}
}

func (s Stage) Next() Stage {
	switch s {
	case StageWait:
		return StageRegistrationOpened
	case StageRegistrationOpened:
		return StageRegistrationClosed
	case StageRegistrationClosed:
		return StageSubmissionOpened
	case StageSubmissionOpened:
		return StageSubmissionClosed
	case StageSubmissionClosed:
		return StagePresentationStarted
	case StagePresentationStarted:
		return StagePresentationFinished
	case StagePresentationFinished:
		return StageReviewOpened
	case StageReviewOpened:
		return StageReviewClosed
	default:
		return StageNone
	}
}
