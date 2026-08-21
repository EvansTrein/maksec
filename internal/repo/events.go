package repo

import (
	"context"
	"fmt"

	"maksec/internal/entity"
	"maksec/pkg/postgres"
	"maksec/pkg/sqldb"

	"github.com/rs/zerolog"
)

const (
	queryEventInsert = "Event.Insert"
)

type Events struct {
	log     *zerolog.Logger
	repo    *postgres.Database
	queries *sqldb.Queries
}

func MustNewEvents(log *zerolog.Logger, repo *postgres.Database) *Events {
	r := &Events{
		log:     log,
		repo:    repo,
		queries: sqldb.NewRequestsDatabase(repo),
	}

	if err := r.init(); err != nil {
		r.log.Fatal().Err(err).Msg("failed to init queries for Events repo")
	}

	return r
}

func (r *Events) init() error {
	if err := r.queries.AddNamedStmtRequest(queryEventInsert, `
		INSERT INTO events (script_path, agent_user, action, event_time)
		VALUES (:script_path, :agent_user, :action, :event_time)
	`); err != nil {
		return err
	}

	return nil
}

func (r *Events) Create(ctx context.Context, event *entity.EventRow) error {
	insertStmt, err := r.queries.NamedStmtRequest(queryEventInsert)
	if err != nil {
		return fmt.Errorf("get insert stmt: %w", err)
	}

	if _, err = insertStmt.ExecContext(ctx, event); err != nil {
		return fmt.Errorf("insert event: %w", err)
	}

	return nil
}
