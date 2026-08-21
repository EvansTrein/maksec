package httpserver

import (
	"context"
	"maksec/internal/entity"
	"maksec/pkg/rest"
	"maksec/pkg/util"
	"net/http"
)

type ICallbackUseCase interface {
	Exec(ctx context.Context, event *entity.Event) error
}

type HandlerCallback struct {
	base *HandlerBase
	us   ICallbackUseCase
}

func NewHandlerCallback(base *HandlerBase, us ICallbackUseCase) *HandlerCallback {
	return &HandlerCallback{
		base: base,
		us:   us,
	}
}

func (h *HandlerCallback) Listen() rest.HandlerFunc {
	return func(c *rest.Context) {
		event, err := util.DecodeBody[entity.Event](c.Request().Body)
		if err != nil {
			h.base.log.Warn().Err(err).Msg("failed to read body")
			h.base.ErrorDetectJSON(c, err)
			return
		}

		if err := h.us.Exec(c.Request().Context(), event); err != nil {
			h.base.log.Error().Err(err).Msg("failed to exec callback")
			h.base.ErrorDetectJSON(c, err)
			return
		}

		if err := c.JSON(http.StatusNoContent, nil); err != nil {
			h.base.log.Error().Err(err).Msg("failed to send json response")
		}
	}
}
