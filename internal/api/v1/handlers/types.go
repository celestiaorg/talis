package handlers

type Slug string

const (
	SuccessSlug      Slug = "success"
	ErrorSlug        Slug = "error"
	InvalidInputSlug Slug = "invalid-input"
	ServerErrorSlug  Slug = "server-error"
)

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

func errGeneral(msg string) Response {
	return Response{
		Slug:  ErrorSlug,
		Error: msg,
	}
}
