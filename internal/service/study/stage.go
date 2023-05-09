package study

type StudyStage uint8

const (
	StudyStageNone                 StudyStage = 0
	StudyStageWait                 StudyStage = 1
	StudyStageRegistrationOpend    StudyStage = 2
	StudyStageRegistrationClosed   StudyStage = 3
	StudyStageSubmissionOpend      StudyStage = 4
	StudyStageSubmissionClosed     StudyStage = 5
	StudyStagePresentationStarted  StudyStage = 6
	StudyStagePresentationFinished StudyStage = 7
	StudyStageReviewOpened         StudyStage = 8
	StudyStageReviewClosed         StudyStage = 9
)

func (s StudyStage) String() string {
	switch s {
	case StudyStageWait:
		return "다음 회차 대기"
	case StudyStageRegistrationOpend:
		return "발표자 등록"
	case StudyStageRegistrationClosed:
		return "발표자 등록 마감"
	case StudyStageSubmissionOpend:
		return "발표 자료 제출"
	case StudyStageSubmissionClosed:
		return "발표 자료 제출 마감"
	case StudyStagePresentationStarted:
		return "발표"
	case StudyStagePresentationFinished:
		return "발표 종료"
	case StudyStageReviewOpened:
		return "피드백"
	case StudyStageReviewClosed:
		return "피드백 마감"
	default:
		return "초기화"
	}
}

func (s StudyStage) IsNone() bool {
	return s == StudyStageNone
}

func (s StudyStage) IsWait() bool {
	return s == StudyStageWait
}

func (s StudyStage) IsRegistrationOpened() bool {
	return s == StudyStageRegistrationOpend
}

func (s StudyStage) IsRegistrationClosed() bool {
	return s == StudyStageRegistrationClosed
}

func (s StudyStage) IsSubmissionOpened() bool {
	return s == StudyStageSubmissionOpend
}

func (s StudyStage) IsSubmissionClosed() bool {
	return s == StudyStageSubmissionClosed
}

func (s StudyStage) IsPresentationStarted() bool {
	return s == StudyStagePresentationStarted
}

func (s StudyStage) IsPresentationFinished() bool {
	return s == StudyStagePresentationFinished
}

func (s StudyStage) IsReviewOpened() bool {
	return s == StudyStageReviewOpened
}

func (s StudyStage) IsReviewClosed() bool {
	return s == StudyStageReviewClosed
}
