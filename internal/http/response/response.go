package response

type Response struct {
	Success bool           `json:"success"`
	Error   *ErrorResponse `json:"error,omitempty"`
}
