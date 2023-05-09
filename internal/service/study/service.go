package study

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrManagementNotFound     = errors.New("스터디 관리 정보를 찾을 수 없습니다.")
	ErrStudyNotFound          = errors.New("스터디 정보를 찾을 수 없습니다.")
	ErrNotManager             = errors.New("스터디 관리자가 아닙니다.")
	ErrStudyExists            = errors.New("이미 진행중인 스터디가 있습니다.")
	ErrInvalidStudyStage      = errors.New("잘못된 스터디 단계입니다.")
	ErrMemberNotFound         = errors.New("스터디 멤버를 찾을 수 없습니다.")
	ErrPresentorNotRegistered = errors.New("발표자로 등록되지 않은 멤버입니다.")
	ErrReviewByYourself       = errors.New("자기 자신을 리뷰할 수 없습니다.")
	ErrEmptyStudyContent      = errors.New("발표 자료 링크가 비어있습니다.")
)

type Service interface {
	GetManagement(ctx context.Context, guildID string) (*Management, error)
	GetOngoingStudy(ctx context.Context, guildID string) (*Study, error)
	GetStudies(ctx context.Context, guildID string) ([]*Study, error)
	InitNewStudyRound(ctx context.Context, guildID, managerID, title string, memberIDs []string) error
	SetMemberRegistered(ctx context.Context, guildID, memberID, name, subject string, registered bool) error
	CloseRegistration(ctx context.Context, guildID, managerID string) error
	OpenSubmission(ctx context.Context, guildID, managerID string) error
	SubmitContent(ctx context.Context, guildID, memberID, contentURL string) error
	CloseSubmission(ctx context.Context, guildID, managerID string) error
	StartPresentation(ctx context.Context, guildID, managerID string) error
	SetPresentorAttended(ctx context.Context, guildID, managerID, memberID string, attended bool) error
	FinishPresentation(ctx context.Context, guildID, managerID string) error
	SetStudyContent(ctx context.Context, guildID, managerID, contentURL string) error
	OpenReview(ctx context.Context, guildID, managerID string) error
	SetReviewer(ctx context.Context, guildID, reviewerID, revieweeID string) error
	CloseReview(ctx context.Context, guildID, managerID string) error
	CloseStudyRound(ctx context.Context, guildID, managerID string) error
	SetNoticeChannelID(ctx context.Context, guildID, managerID, channelID string) error
}

type serviceImpl struct {
	tx Tx

	mtx *sync.Mutex
}

// create new service
func NewService(ctx context.Context, tx Tx, guildID, managerID, noticeChID string) (Service, error) {
	svc := &serviceImpl{
		tx:  tx,
		mtx: &sync.Mutex{},
	}
	return svc.setup(ctx, guildID, managerID, noticeChID)
}

// setup service
func (svc *serviceImpl) setup(ctx context.Context, guildID, managerID, noticeChID string) (*serviceImpl, error) {
	svc.mtx.Lock()
	defer svc.mtx.Unlock()

	// transaction for setup
	txFn := func(sc context.Context) (interface{}, error) {
		// find management
		m, err := svc.tx.FindManagement(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no management, create one
		if m == nil {
			nm := NewManagement()

			nm.SetGuildID(guildID)
			nm.SetManagerID(managerID)
			nm.SetNoticeChannelID(noticeChID)

			id, err := svc.tx.StoreManagement(ctx, nm)
			if err != nil {
				return nil, err
			}

			nm.SetID(id)
		}

		return nil, nil
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	if err != nil {
		return nil, err
	}

	// return serviceImpl
	return svc, nil
}

// get management info of guild
func (svc *serviceImpl) GetManagement(ctx context.Context, guildID string) (*Management, error) {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	m, err := svc.tx.FindManagement(ctx, guildID)
	if err != nil {
		return nil, err
	}

	if m == nil {
		return nil, ErrManagementNotFound
	}

	return m, nil
}

// get ongoing study of guild
func (svc *serviceImpl) GetOngoingStudy(ctx context.Context, guildID string) (*Study, error) {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		// find management
		m, err := svc.tx.FindManagement(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no management, return error
		if m == nil {
			return nil, ErrManagementNotFound
		}

		// if there is no ongoing study, return error
		if m.OngoingStudyID == "" {
			return nil, ErrStudyNotFound
		}

		// find ongoing study
		s, err := svc.tx.FindStudy(sc, m.OngoingStudyID)
		if err != nil {
			return nil, err
		}

		// if there is no ongoing study, return error
		if s == nil {
			return nil, ErrStudyNotFound
		}

		return s, nil
	}

	// execute transaction
	res, err := svc.tx.ExecTx(ctx, txFn)
	if err != nil {
		return nil, err
	}

	return res.(*Study), nil
}

// get all studies of guild
func (svc *serviceImpl) GetStudies(ctx context.Context, guildID string) ([]*Study, error) {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	s, err := svc.tx.FindStudies(ctx, guildID)
	if err != nil {
		return nil, err
	}

	if s == nil {
		return nil, ErrStudyNotFound
	}

	return s, nil
}

// initialize new study round
func (svc *serviceImpl) InitNewStudyRound(ctx context.Context, guildID, managerID, title string, memberIDs []string) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		// find management
		m, err := svc.tx.FindManagement(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no management, return error
		if m == nil {
			return nil, ErrManagementNotFound
		}

		// check if valid manager
		if !m.IsManager(managerID) {
			return nil, ErrNotManager
		}

		// check if there is any ongoing study
		if !(m.CurrentStudyStage.IsNone() || m.CurrentStudyStage.IsWait()) {
			return nil, ErrStudyExists
		}

		// create study
		s := New()
		s.SetGuildID(m.GuildID)
		s.SetTitle(title)

		// set initial members
		for _, id := range memberIDs {
			member := NewMember()
			s.SetMember(id, member)
		}

		// store new study
		studyID, err := svc.tx.StoreStudy(sc, s)
		if err != nil {
			return nil, err
		}

		// set study id to management and move to registration started stage
		m.SetOngoingStudyID(studyID)
		m.SetCurrentStudyStage(StudyStageRegistrationOpend)
		m.SetUpdatedAt(time.Now())

		// update management
		err = svc.tx.UpdateManagement(sc, *m)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// set member's presentation registered status
func (svc *serviceImpl) SetMemberRegistered(ctx context.Context, guildID, memberID, name, subject string, registered bool) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	// transaction for changing member registration
	txFn := func(sc context.Context) (interface{}, error) {
		// find management
		m, err := svc.tx.FindManagement(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no management, return error
		if m == nil {
			return nil, ErrManagementNotFound
		}

		// check if study is in registration stage
		if !m.CurrentStudyStage.IsRegistrationOpened() {
			return nil, errors.Join(ErrInvalidStudyStage, errors.New("발표자 등록 및 등록 해지가 불가능한 단계입니다."))
		}

		// find ongoing study
		s, err := svc.tx.FindStudy(sc, m.OngoingStudyID)
		if err != nil {
			return nil, err
		}

		// if there is no ongoing study, return error
		if s == nil {
			return nil, ErrStudyNotFound
		}

		// check if member is initialized
		member, ok := s.GetMember(memberID)
		if !ok {
			member = NewMember()
		}

		// change member's registered state
		if name != "" {
			member.SetName(name)
		}

		if subject != "" {
			member.SetSubject(subject)
		}

		member.SetRegistered(registered)

		// set updated member to study
		s.SetMember(memberID, member)
		s.SetUpdatedAt(time.Now())

		// update study
		err = svc.tx.UpdateStudy(sc, *s)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}

	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// close registration
func (svc *serviceImpl) CloseRegistration(ctx context.Context, guildID, managerID string) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		// find management
		m, err := svc.tx.FindManagement(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no management, return error
		if m == nil {
			return nil, ErrManagementNotFound
		}

		// check if valid manager
		if !m.IsManager(managerID) {
			return nil, ErrNotManager
		}

		// check if study is in registration stage
		if !m.CurrentStudyStage.IsRegistrationOpened() {
			return nil, errors.Join(ErrInvalidStudyStage, errors.New("발표자 등록 마감이 불가능한 단계입니다."))
		}

		m.SetCurrentStudyStage(StudyStageRegistrationClosed)
		m.SetUpdatedAt(time.Now())

		err = svc.tx.UpdateManagement(sc, *m)
		return nil, err
	}

	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// open submission
func (svc *serviceImpl) OpenSubmission(ctx context.Context, guildID, managerID string) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		// find management
		m, err := svc.tx.FindManagement(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no management, return error
		if m == nil {
			return nil, ErrManagementNotFound
		}

		// check if valid manager
		if !m.IsManager(managerID) {
			return nil, ErrNotManager
		}

		// check if study is in registration closed stage
		if !m.CurrentStudyStage.IsRegistrationClosed() {
			return nil, errors.Join(ErrInvalidStudyStage, errors.New("발표 자료 제출 시작이 불가능한 단계입니다."))
		}

		m.SetCurrentStudyStage(StudyStageSubmissionOpend)
		m.SetUpdatedAt(time.Now())

		err = svc.tx.UpdateManagement(sc, *m)
		return nil, err
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// submit content
func (svc *serviceImpl) SubmitContent(ctx context.Context, guildID, memberID, contentURL string) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		// find management
		m, err := svc.tx.FindManagement(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no management, return error
		if m == nil {
			return nil, ErrManagementNotFound
		}

		// check if study is in submission stage
		if !m.CurrentStudyStage.IsSubmissionOpened() {
			return nil, errors.Join(ErrInvalidStudyStage, errors.New("발표 자료 제출이 불가능한 단계입니다."))
		}

		// find ongoing study
		s, err := svc.tx.FindStudy(sc, m.OngoingStudyID)
		if err != nil {
			return nil, err
		}

		// if there is no ongoing study, return error
		if s == nil {
			return nil, ErrStudyNotFound
		}

		// check if member is initialized
		member, ok := s.GetMember(memberID)
		if !ok {
			return nil, ErrMemberNotFound
		}

		// check if member is registered
		if !member.Registered {
			return nil, ErrPresentorNotRegistered
		}

		// set content
		member.SetContentURL(contentURL)

		// set updated member to study
		s.SetMember(memberID, member)
		s.SetUpdatedAt(time.Now())

		// update study
		err = svc.tx.UpdateStudy(sc, *s)
		return nil, err
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// close submission
func (svc *serviceImpl) CloseSubmission(ctx context.Context, guildID, managerID string) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		m, err := svc.tx.FindManagement(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no management, return error
		if m == nil {
			return nil, ErrManagementNotFound
		}

		// check if valid manager
		if !m.IsManager(managerID) {
			return nil, ErrNotManager
		}

		// check if study is in submission stage
		if !m.CurrentStudyStage.IsSubmissionOpened() {
			return nil, errors.Join(ErrInvalidStudyStage, errors.New("발표 자료 제출 마감이 불가능한 단계입니다."))
		}

		m.SetCurrentStudyStage(StudyStageSubmissionClosed)
		m.SetUpdatedAt(time.Now())

		err = svc.tx.UpdateManagement(sc, *m)
		return nil, err
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// start presentation
func (svc *serviceImpl) StartPresentation(ctx context.Context, guildID, managerID string) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		m, err := svc.tx.FindManagement(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no management, return error
		if m == nil {
			return nil, ErrManagementNotFound
		}

		// check if valid manager
		if !m.IsManager(managerID) {
			return nil, ErrNotManager
		}

		// check if content submission is finished
		if !m.CurrentStudyStage.IsSubmissionClosed() {
			return nil, errors.Join(ErrInvalidStudyStage, errors.New("발표 시작이 불가능한 단계입니다."))
		}

		m.SetCurrentStudyStage(StudyStagePresentationStarted)
		m.SetUpdatedAt(time.Now())

		err = svc.tx.UpdateManagement(sc, *m)
		return nil, err
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// set presentor attended
func (svc *serviceImpl) SetPresentorAttended(ctx context.Context, guildID, managerID, memberID string, attended bool) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		// find management
		m, err := svc.tx.FindManagement(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no management, return error
		if m == nil {
			return nil, ErrManagementNotFound
		}

		// check if valid manager
		if !m.IsManager(managerID) {
			return nil, ErrNotManager
		}

		// check if presentation is started
		if m.CurrentStudyStage < StudyStagePresentationStarted {
			return nil, errors.Join(ErrInvalidStudyStage, errors.New("발표자 출석 확인이 불가능한 단계입니다."))
		}

		// find ongoing study
		s, err := svc.tx.FindStudy(sc, m.OngoingStudyID)
		if err != nil {
			return nil, err
		}

		// if there is no ongoing study, return error
		if s == nil {
			return nil, ErrStudyNotFound
		}

		// check if member is initialized
		member, ok := s.GetMember(memberID)
		if !ok {
			return nil, ErrMemberNotFound
		}

		// check if member is registered
		if !member.Registered {
			return nil, ErrPresentorNotRegistered
		}

		// set attended
		member.SetAttended(attended)

		// set updated member to study
		s.SetMember(memberID, member)
		s.SetUpdatedAt(time.Now())

		// update study
		err = svc.tx.UpdateStudy(sc, *s)
		return nil, err
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// finish presentation
func (svc *serviceImpl) FinishPresentation(ctx context.Context, guildID, managerID string) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		// find management
		m, err := svc.tx.FindManagement(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no management, return error
		if m == nil {
			return nil, ErrManagementNotFound
		}

		// check if valid manager
		if !m.IsManager(managerID) {
			return nil, ErrNotManager
		}

		// check if presentation is ongoing
		if !m.CurrentStudyStage.IsPresentationStarted() {
			return nil, errors.Join(ErrInvalidStudyStage, errors.New("발표 종료가 불가능한 단계입니다."))
		}

		m.SetCurrentStudyStage(StudyStagePresentationFinished)
		m.SetUpdatedAt(time.Now())

		err = svc.tx.UpdateManagement(sc, *m)
		return nil, err
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// set study content
func (svc *serviceImpl) SetStudyContent(ctx context.Context, guildID, managerID, content string) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		// find management
		m, err := svc.tx.FindManagement(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no management, return error
		if m == nil {
			return nil, ErrManagementNotFound
		}

		// check if valid manager
		if !m.IsManager(managerID) {
			return nil, ErrNotManager
		}

		// check if presentation is finished
		if m.CurrentStudyStage < StudyStagePresentationFinished {
			return nil, errors.Join(ErrInvalidStudyStage, errors.New("스터디 자료 링크 등록이 불가능한 단계입니다."))
		}

		// find ongoing study
		s, err := svc.tx.FindStudy(sc, m.OngoingStudyID)
		if err != nil {
			return nil, err
		}

		// if there is no ongoing study, return error
		if s == nil {
			return nil, ErrStudyNotFound
		}

		if content == "" {
			return nil, ErrEmptyStudyContent
		}

		// set content
		s.SetContentURL(content)
		s.SetUpdatedAt(time.Now())

		// update study
		err = svc.tx.UpdateStudy(sc, *s)
		return nil, err
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// open review
func (svc *serviceImpl) OpenReview(ctx context.Context, guildID, managerID string) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		// find management
		m, err := svc.tx.FindManagement(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no management, return error
		if m == nil {
			return nil, ErrManagementNotFound
		}

		// check if valid manager
		if !m.IsManager(managerID) {
			return nil, ErrNotManager
		}

		// check if presentation is finished
		if !m.CurrentStudyStage.IsPresentationFinished() {
			return nil, errors.Join(ErrInvalidStudyStage, errors.New("리뷰 시작이 불가능한 단계입니다."))
		}

		// update management
		m.SetCurrentStudyStage(StudyStageReviewOpened)
		m.SetUpdatedAt(time.Now())

		err = svc.tx.UpdateManagement(sc, *m)
		return nil, err
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// set reviewer
func (svc *serviceImpl) SetReviewer(ctx context.Context, guildID, reviewerID, revieweeID string) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	if reviewerID == revieweeID {
		return ErrReviewByYourself
	}

	txFn := func(sc context.Context) (interface{}, error) {
		// find management
		m, err := svc.tx.FindManagement(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no management, return error
		if m == nil {
			return nil, ErrManagementNotFound
		}

		// check if valid reviewer
		if !m.IsManager(reviewerID) {
			return nil, ErrNotManager
		}

		// check if review is ongoing
		if !m.CurrentStudyStage.IsReviewOpened() {
			return nil, errors.Join(ErrInvalidStudyStage, errors.New("리뷰 작성이 불가능한 단계입니다."))
		}

		// find ongoing study
		s, err := svc.tx.FindStudy(sc, m.OngoingStudyID)
		if err != nil {
			return nil, err
		}

		// if there is no ongoing study, return error
		if s == nil {
			return nil, ErrStudyNotFound
		}

		// check if reviewer is member of ongoing study
		_, ok := s.GetMember(reviewerID)
		if !ok {
			return nil, errors.Join(ErrMemberNotFound, errors.New("리뷰어는 스터디에 참여한 사용자여야 합니다."))
		}

		// check if reviewee is member of ongoing study
		reviewee, ok := s.GetMember(revieweeID)
		if !ok {
			return nil, errors.Join(ErrMemberNotFound, errors.New("리뷰 대상자는 스터디에 참여한 사용자여야 합니다."))
		}

		// check if reviewee is registered and attended presentation
		if !reviewee.Registered || !reviewee.Attended {
			return nil, errors.New("리뷰 대상자는 발표에 참여한 사용자여야 합니다.")
		}

		// check if reviewer already reviewed
		if reviewee.IsReviewer(reviewerID) {
			return nil, errors.New("이미 리뷰를 작성하였습니다.")
		}

		// set reviewer
		reviewee.SetReviewer(reviewerID)

		// set updated member to study
		s.SetMember(revieweeID, reviewee)
		s.SetUpdatedAt(time.Now())

		// update study
		err = svc.tx.UpdateStudy(sc, *s)
		return nil, err
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// close review
func (svc *serviceImpl) CloseReview(ctx context.Context, guildID, managerID string) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		// find management
		m, err := svc.tx.FindManagement(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no management, return error
		if m == nil {
			return nil, ErrManagementNotFound
		}

		// check if valid manager
		if !m.IsManager(managerID) {
			return nil, ErrNotManager
		}

		// check if review is ongoing
		if !m.CurrentStudyStage.IsReviewOpened() {
			return nil, errors.Join(ErrInvalidStudyStage, errors.New("리뷰 작성 마감이 불가능한 단계입니다."))
		}

		// find ongoing study
		s, err := svc.tx.FindStudy(sc, m.OngoingStudyID)
		if err != nil {
			return nil, err
		}

		// if there is no ongoing study, return error
		if s == nil {
			return nil, ErrStudyNotFound
		}

		// update management
		m.SetCurrentStudyStage(StudyStageReviewClosed)
		m.SetUpdatedAt(time.Now())

		err = svc.tx.UpdateManagement(sc, *m)
		return nil, err
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// close study round
func (svc *serviceImpl) CloseStudyRound(ctx context.Context, guildID, managerID string) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		// find management
		m, err := svc.tx.FindManagement(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no management, return error
		if m == nil {
			return nil, ErrManagementNotFound
		}

		// check if valid manager
		if !m.IsManager(managerID) {
			return nil, ErrNotManager
		}

		// check if review is finished
		if !m.CurrentStudyStage.IsReviewClosed() {
			return nil, errors.Join(ErrInvalidStudyStage, errors.New("스터디 라운드 종료가 불가능한 단계입니다."))
		}

		// update management
		m.SetOngoingStudyID("")
		m.SetCurrentStudyStage(StudyStageWait)
		m.SetUpdatedAt(time.Now())

		err = svc.tx.UpdateManagement(sc, *m)
		return nil, err
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// set notice channel id
func (svc *serviceImpl) SetNoticeChannelID(ctx context.Context, guildID, managerID, channelID string) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		// find management
		m, err := svc.tx.FindManagement(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no management, return error
		if m == nil {
			return nil, ErrManagementNotFound
		}

		// check if valid manager
		if !m.IsManager(managerID) {
			return nil, ErrNotManager
		}

		// update management
		m.SetNoticeChannelID(channelID)
		m.SetUpdatedAt(time.Now())

		err = svc.tx.UpdateManagement(sc, *m)
		return nil, err
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}
