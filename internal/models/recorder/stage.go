package recorder

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
