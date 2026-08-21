package rest

import "strings"

// Problem — тело ошибки для REST-ответов (application/problem+json).
// service — расширение: сервис, вернувший ошибку.
type Problem struct {
	Type    string `json:"type"`
	Status  int    `json:"status"`
	Detail  string `json:"detail,omitempty"`
	Service string `json:"service,omitempty"`
}

// ProblemFromError строит Problem из строки ошибки формата "code: description".
func ProblemFromError(status int, text, service string) Problem {
	code, detail, ok := strings.Cut(text, ": ")
	if !ok {
		code = "error"
		detail = text
	}
	return Problem{Type: code, Status: status, Detail: detail, Service: service}
}
