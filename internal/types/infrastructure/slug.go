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
	NotFoundSlug     Slug = "not-found"
)

// SlugResponse is the response type for the API
// TODO: I think this needs to be revised, I think a generic APIResponse type would be better. The Client methods would need to be updated to ingest the APIResponse type and then decode the data into the appropriate type.
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

// ErrNotFound returns a SlugResponse with the NotFoundSlug and the error message
func ErrNotFound(msg string) SlugResponse {
	return SlugResponse{
		Slug:  NotFoundSlug,
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

// ErrNotFound returns a SlugResponse with the NotFoundSlug and error message
func ErrNotFound(msg string) SlugResponse {
	return SlugResponse{
		Slug:  NotFoundSlug,
		Error: msg,
	}
}
