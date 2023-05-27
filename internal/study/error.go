package study

import "errors"

var (
	ErrStudyExists           = errors.New("이미 진행중인 스터디가 있습니다")
	ErrStudyNotFound         = errors.New("스터디 정보를 찾을 수 없습니다")
	ErrRoundNotFound         = errors.New("라운드 정보를 찾을 수 없습니다")
	ErrInvalidStage          = errors.New("잘못된 스터디 단계입니다")
	ErrAlreadyRegistered     = errors.New("이미 등록된 발표자입니다")
	ErrAlreadyUnregistered   = errors.New("이미 등록 해제된 발표자입니다")
	ErrNotRegistered         = errors.New("등록되지 않은 발표자입니다")
	ErrNotAttended           = errors.New("참석하지 않은 발표자입니다")
	ErrMemberNotFound        = errors.New("등록된 사용자 정보를 찾을 수 없습니다")
	ErrReviewByYourself      = errors.New("자기 자신을 리뷰할 수 없습니다")
	ErrAlreadySentReflection = errors.New("이미 회고를 작성하셨습니다")
	ErrNilUpdateParams       = errors.New("업데이트할 정보가 없습니다")
	ErrInvalidUpdateParams   = errors.New("잘못된 업데이트 파라미터입니다")
	ErrAlreadySentReview     = errors.New("이미 리뷰를 작성하셨습니다")
)
