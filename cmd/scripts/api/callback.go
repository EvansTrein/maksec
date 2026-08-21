package api

import (
	"maksec/internal/adapter/httpserver"
	"maksec/pkg/rest"
	"net/http"
)

type CallbackActiveHandlers struct {
	callback *httpserver.HandlerCallback
}

func NewCallbackActiveHandlers(callback *httpserver.HandlerCallback) *CallbackActiveHandlers {
	return &CallbackActiveHandlers{callback: callback}
}

func InitRouterCallback(version, service string, handlers *CallbackActiveHandlers) *rest.Router {
	r := rest.NewRouter(
		version,
		service,
	)

	r.AddRoute(&rest.Route{
		Method:  http.MethodPost,
		Path:    "callback",
		Handler: handlers.callback.Listen(),
		// Middlewares: , // auth etc.
	})

	return r
}
