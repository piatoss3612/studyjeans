package study

type Member struct {
	Name           string          `bson:"name" json:"name"`
	Subject        string          `bson:"subject" json:"subject"`
	ContentURL     string          `bson:"content_url" json:"content_url"`
	Registered     bool            `bson:"registered" json:"registered"`
	Attended       bool            `bson:"attended" json:"attended"`
	SentReflection bool            `bson:"sent_reflection" json:"sent_reflection,omitempty"`
	Reviewers      map[string]bool `bson:"reviewers" json:"reviewers,omitempty"`
}

func NewMember() Member {
	return Member{
		Name:       "",
		Subject:    "",
		ContentURL: "",
		Registered: false,
		Attended:   false,
		Reviewers:  map[string]bool{},
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

func (m *Member) SetAttended(Attended bool) {
	m.Attended = Attended
}

func (m *Member) SetSentReflection(sentReflection bool) {
	m.SentReflection = sentReflection
}

func (m *Member) SetReviewer(userID string) {
	m.Reviewers[userID] = true
}

func (m *Member) IsReviewer(userID string) bool {
	return m.Reviewers[userID]
}
