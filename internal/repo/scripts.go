package repo

import (
	"context"
	"fmt"

	"maksec/internal/entity"
	"maksec/pkg/postgres"
	"maksec/pkg/sqldb"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

const (
	queryScriptUpsert  = "Script.Upsert"
	queryScriptGetByID = "Script.Get.ID"
)

type Scripts struct {
	log     *zerolog.Logger
	repo    *postgres.Database
	queries *sqldb.Queries
}

func MustNewScripts(log *zerolog.Logger, repo *postgres.Database) *Scripts {
	r := &Scripts{
		log:     log,
		repo:    repo,
		queries: sqldb.NewRequestsDatabase(repo),
	}

	if err := r.init(); err != nil {
		r.log.Fatal().Err(err).Msg("failed to init queries for Scripts repo")
	}

	return r
}

func (r *Scripts) init() error {
	if err := r.queries.AddNamedStmtRequest(queryScriptUpsert, `
		INSERT INTO scripts (host, ssh_user, password, template, path)
		VALUES (:host, :ssh_user, :password, :template, :path)
		ON CONFLICT (host, path) DO UPDATE SET
			created_at = NOW()
		RETURNING id
	`); err != nil {
		return err
	}

	if err := r.queries.AddStmtRequest(queryScriptGetByID, `
		SELECT id, host, ssh_user, password, template, path, created_at
		FROM scripts 
		WHERE id = $1
	`); err != nil {
		return err
	}

	return nil
}

func (r *Scripts) Create(ctx context.Context, script *entity.Script) (*entity.Script, error) {
	var (
		tx  *sqlx.Tx
		id  int64
		err error
	)

	if tx, err = r.repo.Connection().Beginx(); err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				r.log.Error().Err(rbErr).Msg("failed to rollback tx")
			}
		}
	}()

	upsertStmt, err := r.queries.NamedStmtRequest(queryScriptUpsert)
	if err != nil {
		return nil, fmt.Errorf("get insert stmt: %w", err)
	}

	if err = tx.NamedStmtContext(ctx, upsertStmt).QueryRowContext(ctx, script).Scan(&id); err != nil {
		return nil, fmt.Errorf("insert script: %w", err)
	}

	getStmt, err := r.queries.StmtRequest(queryScriptGetByID)
	if err != nil {
		return nil, fmt.Errorf("get by id stmt: %w", err)
	}

	if err = tx.StmtxContext(ctx, getStmt).GetContext(ctx, script, id); err != nil {
		return nil, fmt.Errorf("reload script %d: %w", id, err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return script, nil
}
