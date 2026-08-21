package rest

import (
	"encoding/json"
	"net/http"
)

type Context struct {
	w       http.ResponseWriter
	r       *http.Request
	written bool
}

func (c *Context) JSON(code int, v any) error {
	if c.written {
		return nil
	}
	c.written = true

	c.w.Header().Set("Content-Type", "application/json")
	c.w.WriteHeader(code)
	return json.NewEncoder(c.w).Encode(v)
}

// Problem рендерит ошибку в формате application/problem+json.
func (c *Context) Problem(problem Problem) error {
	if c.written {
		return nil
	}
	c.written = true

	c.w.Header().Set("Content-Type", "application/problem+json")
	c.w.WriteHeader(problem.Status)
	return json.NewEncoder(c.w).Encode(problem)
}

func (c *Context) Request() *http.Request {
	return c.r
}
