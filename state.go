package main

type StudyState uint8

const (
	StudyStateNone     StudyState = 0
	StudyStateWait     StudyState = 1
	StudyStateRegister StudyState = 2
	StudyStateSubmit   StudyState = 3
	StudyStatePresent  StudyState = 4
	StudyStateReview   StudyState = 5
)

func (s StudyState) String() string {
	switch s {
	case StudyStateWait:
		return "다음 주차 대기"
	case StudyStateRegister:
		return "발표자 등록"
	case StudyStateSubmit:
		return "발표자료 제출"
	case StudyStatePresent:
		return "발표"
	case StudyStateReview:
		return "리뷰 및 피드백"
	default:
		return "몰?루"
	}
}
