package server

type apiError struct {
	Message string `json:"error"`
	Code    int    `json:"code"`
}

func newAPIError(msg string, code int) apiError {
	return apiError{
		Message: msg,
		Code:    code,
	}
}

func (e apiError) Error() string {
	return e.Message
}
