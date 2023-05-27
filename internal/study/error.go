package study

import "errors"

var (
	ErrStudyExists           = errors.New("이미 진행중인 스터디가 있습니다")
	ErrRoundExists           = errors.New("이미 진행중인 라운드가 있습니다")
	ErrInvalidManager        = errors.New("잘못된 매니저입니다")
	ErrStudyNotFound         = errors.New("스터디 정보를 찾을 수 없습니다")
	ErrRoundNotFound         = errors.New("라운드 정보를 찾을 수 없습니다")
	ErrInvalidStage          = errors.New("잘못된 스터디 단계입니다")
	ErrAlreadyRegistered     = errors.New("이미 등록된 발표자입니다")
	ErrAlreadyUnregistered   = errors.New("이미 등록 해제된 발표자입니다")
	ErrMemberNotRegistered   = errors.New("등록되지 않은 발표자입니다")
	ErrMemberNotAttended     = errors.New("참석하지 않은 발표자입니다")
	ErrMemberNotFound        = errors.New("등록된 사용자 정보를 찾을 수 없습니다")
	ErrReviewByYourself      = errors.New("자기 자신을 리뷰할 수 없습니다")
	ErrAlreadySentReflection = errors.New("이미 회고를 작성하셨습니다")
	ErrNilParams             = errors.New("파라미터가 nil입니다")
	ErrInvalidUpdateParams   = errors.New("잘못된 업데이트 파라미터입니다")
	ErrAlreadySentReview     = errors.New("이미 리뷰를 작성하셨습니다")
	ErrManagerNotFound       = errors.New("매니저 정보를 찾을 수 없습니다")
	ErrNotManager            = errors.New("매니저만 사용할 수 있는 명령어입니다")
	ErrUserNotFound          = errors.New("사용자 정보를 찾을 수 없습니다")
	ErrChannelNotFound       = errors.New("채널 정보를 찾을 수 없습니다")
	ErrRequiredArgs          = errors.New("필수 인자가 없습니다")
	ErrInvalidArgs           = errors.New("인자가 올바르지 않습니다")
	ErrInvalidCommand        = errors.New("올바르지 않은 명령어입니다")
	ErrRoundAlreadySet       = errors.New("이미 진행중인 스터디 라운드가 있습니다")
	ErrFeedbackYourself      = errors.New("자기 자신에게 피드백을 보낼 수 없습니다")
	ErrNilFunc               = errors.New("함수가 nil입니다")
)
