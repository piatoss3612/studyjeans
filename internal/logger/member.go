package logger

type Member struct {
	Name       string `json:"name"`
	Subject    string `json:"subject"`
	ContentURL string `json:"content_url"`
	Registered bool   `json:"registered"`
	Attended   bool   `json:"attended"`
}

func NewMember() Member {
	return Member{
		Name:       "",
		Subject:    "",
		ContentURL: "",
		Registered: false,
		Attended:   false,
	}
}
