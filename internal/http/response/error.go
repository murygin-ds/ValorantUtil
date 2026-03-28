package response

type ErrorResponse struct {
	Message string      `json:"message,omitempty"`
	Details interface{} `json:"details,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}
