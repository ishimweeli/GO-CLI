package utils

type AssignmentConfig struct {
	ProjectName    string
	Directory      string
	StudentID      string
	AssignmentCode string
}

type ServerError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Status  string `json:"status"`
}
