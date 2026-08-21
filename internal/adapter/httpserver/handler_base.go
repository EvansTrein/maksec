package httpserver

import (
	"errors"
	"net/http"

	"maksec/pkg/apperr"
	"maksec/pkg/rest"

	"github.com/rs/zerolog"
)

type HandlerBase struct {
	log *zerolog.Logger
}

func NewHandlerBase(log *zerolog.Logger) *HandlerBase {
	return &HandlerBase{log: log}
}

func (h *HandlerBase) ErrorDetectJSON(c *rest.Context, err error) {
	status := http.StatusInternalServerError

	switch {
	case errors.Is(err, apperr.ErrRestInvalidBody):
		status = http.StatusBadRequest
	case errors.Is(err, apperr.ErrTemplateNotFound):
		status = http.StatusNotFound
	case errors.Is(err, apperr.ErrSSHAuthFailed):
		status = http.StatusUnauthorized
	}

	if err := c.JSON(status, rest.ProblemFromError(status, err.Error(), "")); err != nil {
		h.log.Error().Err(err).Msg("failed to send json response")
		return
	}
}
