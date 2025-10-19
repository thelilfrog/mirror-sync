package project

type (
	Project struct {
		Name         string       `json:"name"`
		Repositories []Repository `json:"repositories"`
	}

	Repository struct {
		Name        string `json:"name"`
		Schedule    string `json:"schedule"`
		Source      string `json:"source"`
		Destination string `json:"destination"`
	}
)
