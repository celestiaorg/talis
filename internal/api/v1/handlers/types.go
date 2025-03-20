package handlers

// Slug is a type for the slug field in the response
// It is mainly used for the client to understand the type of the response
type Slug string

// nolint:gochecknoglobals
const (
	SuccessSlug      Slug = "success"
	ErrorSlug        Slug = "error"
	InvalidInputSlug Slug = "invalid-input"
	ServerErrorSlug  Slug = "server-error"
)

// Response is the response type for the API
type Response struct {
	Slug  Slug        `json:"slug"`
	Error string      `json:"error"`
	Data  interface{} `json:"data"`
}

func errInvalidInput(msg string) Response {
	return Response{
		Slug:  InvalidInputSlug,
		Error: msg,
	}
}

func errServer(msg string) Response {
	return Response{
		Slug:  ServerErrorSlug,
		Error: msg,
	}
}
