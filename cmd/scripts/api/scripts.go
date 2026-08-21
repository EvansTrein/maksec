package api

import (
	"maksec/internal/adapter/httpserver"
	"maksec/pkg/rest"
	"net/http"
)

type ScriptsActiveHandlers struct {
	scripts *httpserver.HandlerScripts
}

func NewScriptsActiveHandlers(scripts *httpserver.HandlerScripts) *ScriptsActiveHandlers {
	return &ScriptsActiveHandlers{scripts: scripts}
}

func InitRouterScripts(version, service string, handlers *ScriptsActiveHandlers) *rest.Router {
	r := rest.NewRouter(
		version,
		service,
	)

	r.AddRoute(&rest.Route{
		Method:  http.MethodPost,
		Path:    "create",
		Handler: handlers.scripts.Create(),
		// Middlewares: , // auth etc.
	})

	return r
}
