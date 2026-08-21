package usecase

import (
	"context"
	"fmt"

	"maksec/internal/entity"

	"github.com/rs/zerolog"
)

type IRepoEvents interface {
	Create(ctx context.Context, event *entity.EventRow) error
}

type Callback struct {
	log  *zerolog.Logger
	repo IRepoEvents
}

func NewCallback(log *zerolog.Logger, repo IRepoEvents) *Callback {
	return &Callback{log: log, repo: repo}
}

func (us *Callback) Exec(ctx context.Context, event *entity.Event) error {
	log := us.log.With().Str("operation", "usecase-callback").Logger()
	log.Info().Any("event", event).Msg("callback!")

	row := &entity.EventRow{
		ScriptPath: event.ScriptPath,
		AgentUser:  event.User,
		Action:     event.Action,
		EventTime:  event.Time,
	}

	if err := us.repo.Create(ctx, row); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	log.Debug().Msg("success")
	return nil
}
