package response

// Response is the closed set of outcomes a query executor can produce.
type Response interface {
	IsResponse()
}

// ErrorResponse is a Response that already carries the HTTP status to report.
type ErrorResponse struct {
	Status  int
	Message string
}

func (ErrorResponse) IsResponse() {}
