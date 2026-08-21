package httpserver

import (
	"context"
	"maksec/internal/dto"
	"maksec/pkg/rest"
	"maksec/pkg/util"
	"net/http"
)

type IScriptsUseCase interface {
	Create(ctx context.Context, req *dto.ScriptsCreateRequest) (*dto.ScriptsCreateResponse, error)
}

type HandlerScripts struct {
	base *HandlerBase
	us   IScriptsUseCase
}

func NewHandlerScripts(base *HandlerBase, us IScriptsUseCase) *HandlerScripts {
	return &HandlerScripts{
		base: base,
		us:   us,
	}
}

func (h *HandlerScripts) Create() rest.HandlerFunc {
	return func(c *rest.Context) {
		req, err := util.DecodeBody[dto.ScriptsCreateRequest](c.Request().Body)
		if err != nil {
			h.base.log.Warn().Err(err).Msg("failed to read body")
			h.base.ErrorDetectJSON(c, err)
			return
		}

		resp, err := h.us.Create(c.Request().Context(), req)
		if err != nil {
			h.base.log.Error().Err(err).Msg("failed to create script on remote host")
			h.base.ErrorDetectJSON(c, err)
			return
		}

		if err := c.JSON(http.StatusCreated, resp); err != nil {
			h.base.log.Error().Err(err).Msg("failed to send json response")
		}
	}
}
