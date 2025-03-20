package infrastructure

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

// SlugResponse is the response type for the API
type SlugResponse struct {
	Slug  Slug        `json:"slug"`
	Error string      `json:"error"`
	Data  interface{} `json:"data"`
}

// ErrInvalidInput returns a SlugResponse with the InvalidInputSlug and the error message
func ErrInvalidInput(msg string) SlugResponse {
	return SlugResponse{
		Slug:  InvalidInputSlug,
		Error: msg,
	}
}

// ErrServer returns a SlugResponse with the ServerErrorSlug and the error message
func ErrServer(msg string) SlugResponse {
	return SlugResponse{
		Slug:  ServerErrorSlug,
		Error: msg,
	}
}

// Success returns a SlugResponse with the SuccessSlug and the data
func Success(data interface{}) SlugResponse {
	return SlugResponse{
		Slug: SuccessSlug,
		Data: data,
	}
}
