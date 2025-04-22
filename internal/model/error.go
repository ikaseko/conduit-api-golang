package model

type ErrorJson struct {
	Errors struct {
		Body []string `json:"body,omitempty"`
	} `json:"errors"`
}

func NewError(args ...string) *ErrorJson {
	var err ErrorJson
	err.Errors.Body = args
	return &err
}
