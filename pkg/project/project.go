package project

type (
	Project struct {
		UUID         string       `json:"uuid"`
		Name         string       `json:"name"`
		Repositories []Repository `json:"repositories"`
		ServerURL    string       `json:"-"`
	}

	Repository struct {
		UUID            string                            `json:"uuid"`
		Name            string                            `json:"name"`
		Schedule        string                            `json:"schedule"`
		Source          string                            `json:"source"`
		Destination     string                            `json:"destination"`
		Authentications map[string]AuthenticationSettings `json:"authentications"`
	}

	AuthenticationSettings struct {
		Basic *BasicAuthenticationSettings `json:"basic,omitempty"`
		Token string                       `json:"token,omitempty"`
	}

	BasicAuthenticationSettings struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
)
