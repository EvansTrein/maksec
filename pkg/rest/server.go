package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"maksec/internal/config"
	"maksec/pkg/logger"
)

type (
	Option      func(s *Server)
	CheckerFunc func() error
)

type Server struct {
	host string
	port string

	h *http.Server
	r *Router

	logger   *zerolog.Logger
	checkers []CheckerFunc
	ctx      context.Context
}

func MustNewServer(opts ...Option) *Server {
	s := &Server{
		host:     "localhost",
		port:     "8080",
		logger:   &logger.DefaultLogger,
		checkers: make([]CheckerFunc, 0),
		ctx:      config.DefaultCtx(),
	}

	for _, o := range opts {
		o(s)
	}

	if s.r == nil {
		s.logger.Fatal().Msg("router is required")
	}

	mux := http.NewServeMux()

	if s.r.assets != nil {
		for _, asset := range s.r.assets {
			mux.Handle(asset.Prefix, asset.Handler)
		}
	}

	for _, route := range s.r.routes { // сборка всех эндпоинтов
		pattern := fmt.Sprintf("%s %s", route.Method, s.r.pattern(route))
		handler := applyMiddlewares(s.handle(route), route.Middlewares)
		mux.Handle(pattern, handler)
	}

	s.h = &http.Server{
		Addr:    fmt.Sprintf("%s:%s", s.host, s.port),
		Handler: LogMiddleware(s.logger)(mux),
	}

	mux.HandleFunc("/healthz", s.healthz)

	return s
}

// handle адаптирует HandlerFunc к http.Handler: строит Context из запроса.
// Хендлер обязан сам записать ответ (c.JSON/c.Problem) во всех ветках;
// если он завершился, не записав ничего, клиент вместо пустого 200
// получает 500 — это защита от забытого ответа.
func (s *Server) handle(route *Route) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := &Context{w: w, r: r}
		route.Handler(c)
		if !c.written {
			s.logger.Error().Str("method", r.Method).Str("path", r.URL.Path).Msg("handler finished without response")
			_ = c.Problem(Problem{Type: "internal.error", Status: http.StatusInternalServerError})
		}
	})
}

// applyMiddlewares оборачивает handler в цепочку middleware:
// первый элемент — самый внешний, исполняется первым и через next
// передаёт управление дальше либо прерывает обработку.
func applyMiddlewares(handler http.Handler, middlewares []func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

func WithHost(host string) Option {
	return func(s *Server) {
		s.host = host
	}
}

func WithPort(port string) Option {
	return func(s *Server) {
		s.port = port
	}
}

func WithChecker(checkers ...CheckerFunc) Option {
	return func(s *Server) {
		s.checkers = checkers
	}
}

func WithLogger(logger *zerolog.Logger) Option {
	return func(s *Server) {
		s.logger = logger
	}
}

func WithRouter(r *Router) Option {
	return func(s *Server) {
		s.r = r
	}
}

func WithContext(ctx context.Context) Option {
	return func(s *Server) {
		s.ctx = ctx
	}
}

func (s *Server) MustStart() {
	s.logRoutes()
	s.logger.Info().Msgf("http listen on %s", s.h.Addr)
	if err := s.h.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Fatal().Err(err).Msg("failed to start http server")
	}
}

func (s *Server) logRoutes() {
	if s.r == nil {
		return
	}
	for _, route := range s.r.GetRoutes() {
		pattern := fmt.Sprintf("%s http://%s%s", route.Method, s.h.Addr, s.r.pattern(route))
		s.logger.Info().Str("route", pattern).Msg("registered route")
	}
}

func (s *Server) Stop() error {
	if s.h == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	return s.h.Shutdown(ctx)
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	for _, checker := range s.checkers {
		if err := checker(); err != nil {
			s.logger.Error().Err(err).Msg("health check failed")
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
