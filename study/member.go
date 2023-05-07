package study

type Member struct {
	Name         string          `bson:"name"`
	Subject      string          `bson:"subject"`
	ContentURL   string          `bson:"content_url"`
	Registered   bool            `bson:"registered"`
	Participated bool            `bson:"participated"`
	Reviewers    map[string]bool `bson:"reviewers"`
}

func NewMember() Member {
	return Member{
		Name:         "",
		Subject:      "",
		ContentURL:   "",
		Registered:   false,
		Participated: false,
		Reviewers:    map[string]bool{},
	}
}

func (m *Member) SetName(name string) {
	m.Name = name
}

func (m *Member) SetSubject(subject string) {
	m.Subject = subject
}

func (m *Member) SetContentURL(contentURL string) {
	m.ContentURL = contentURL
}

func (m *Member) SetRegistered(registered bool) {
	m.Registered = registered
}

func (m *Member) SetParticipated(Participated bool) {
	m.Participated = Participated
}

func (m *Member) SetReviewer(userID string) {
	m.Reviewers[userID] = true
}

func (m *Member) IsReviewer(userID string) bool {
	return m.Reviewers[userID]
}
