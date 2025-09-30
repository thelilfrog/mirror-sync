package obj

import "time"

type (
	HTTPCore struct {
		Status    int       `json:"status"`
		Timestamp time.Time `json:"timestamp"`
		Path      string    `json:"path"`
	}

	HTTPError struct {
		HTTPCore
		Error   string `json:"error"`
		Message string `json:"message"`
	}

	HTTPObject struct {
		HTTPCore
		Data any `json:"data"`
	}

	SystemInformation struct {
		Version        string `json:"version"`
		APIVersion     int    `json:"api_version"`
		GoVersion      string `json:"go_version"`
		OSName         string `json:"os_name"`
		OSArchitecture string `json:"os_architecture"`
	}
)
