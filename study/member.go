package study

import "sync"

type Member struct {
	Name         string
	Subject      string
	ContentURL   string
	Registered   bool
	Participated bool
	Reviewers    map[string]bool

	mtx *sync.RWMutex
}

func NewMember() Member {
	return Member{
		Name:         "",
		Subject:      "",
		ContentURL:   "",
		Registered:   false,
		Participated: false,
		Reviewers:    map[string]bool{},
		mtx:          &sync.RWMutex{},
	}
}

func (m *Member) SetName(name string) {
	defer m.mtx.Unlock()
	m.mtx.Lock()

	m.Name = name
}

func (m *Member) SetSubject(subject string) {
	defer m.mtx.Unlock()
	m.mtx.Lock()

	m.Subject = subject
}

func (m *Member) SetContentURL(contentURL string) {
	defer m.mtx.Unlock()
	m.mtx.Lock()

	m.ContentURL = contentURL
}

func (m *Member) SetRegistered(registered bool) {
	defer m.mtx.Unlock()
	m.mtx.Lock()

	m.Registered = registered
}

func (m *Member) SetParticipated(Participated bool) {
	defer m.mtx.Unlock()
	m.mtx.Lock()

	m.Participated = Participated
}

func (m *Member) SetReviewer(userID string) {
	defer m.mtx.Unlock()
	m.mtx.Lock()

	m.Reviewers[userID] = true
}

func (m *Member) IsReviewer(userID string) bool {
	defer m.mtx.RUnlock()
	m.mtx.RLock()

	return m.Reviewers[userID]
}
