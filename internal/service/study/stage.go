package study

type StudyStage uint8

const (
	StudyStageNone                 StudyStage = 0
	StudyStageWait                 StudyStage = 1
	StudyStageRegistrationStarted  StudyStage = 2
	StudyStageRegistrationFinished StudyStage = 3
	StudyStageSubmissionStarted    StudyStage = 4
	StudyStageSubmissionFinished   StudyStage = 5
	StudyStagePresentationStarted  StudyStage = 6
	StudyStagePresentationFinished StudyStage = 7
	StudyStageReviewStarted        StudyStage = 8
	StudyStageReviewFinished       StudyStage = 9
)

func (s StudyStage) String() string {
	switch s {
	case StudyStageWait:
		return "다음 회차 대기"
	case StudyStageRegistrationStarted, StudyStageRegistrationFinished:
		return "발표자 등록"
	case StudyStageSubmissionStarted, StudyStageSubmissionFinished:
		return "발표자료 제출"
	case StudyStagePresentationStarted, StudyStagePresentationFinished:
		return "발표"
	case StudyStageReviewStarted, StudyStageReviewFinished:
		return "리뷰 및 피드백"
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

func (s StudyStage) IsRegistrationOngoing() bool {
	return s == StudyStageRegistrationStarted
}

func (s StudyStage) IsRegistrationFinished() bool {
	return s == StudyStageRegistrationFinished
}

func (s StudyStage) IsSubmissionOngoing() bool {
	return s == StudyStageSubmissionStarted
}

func (s StudyStage) IsSubmissionFinished() bool {
	return s == StudyStageSubmissionFinished
}

func (s StudyStage) IsPresentationOngoing() bool {
	return s == StudyStagePresentationStarted
}

func (s StudyStage) IsPresentationFinished() bool {
	return s == StudyStagePresentationFinished
}

func (s StudyStage) IsReviewOngoing() bool {
	return s == StudyStageReviewStarted
}

func (s StudyStage) IsReviewFinished() bool {
	return s == StudyStageReviewFinished
}
