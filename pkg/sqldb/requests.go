package sqldb

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type IDatabase interface {
	Connection() *sqlx.DB
}

type queryName string

type Queries struct {
	namedStmtRequestList map[queryName]NamedStmtRequest
	stmtRequestList      map[queryName]StmtRequest
	db                   IDatabase
}

type NamedStmtRequest struct {
	NamedStmt *sqlx.NamedStmt
	Request   string
}

type StmtRequest struct {
	Stmt    *sqlx.Stmt
	Request string
}

func NewRequestsDatabase(db IDatabase) *Queries {
	return &Queries{
		namedStmtRequestList: make(map[queryName]NamedStmtRequest),
		stmtRequestList:      make(map[queryName]StmtRequest),
		db:                   db,
	}
}

func (r *Queries) AddNamedStmtRequest(queryName queryName, query string) (err error) {
	entry := r.namedStmtRequestList[queryName]
	entry.NamedStmt, err = r.db.Connection().PrepareNamed(query)
	if err != nil {
		return fmt.Errorf("%s failed to prepare query, %w", queryName, err)
	}

	entry.Request = query
	r.namedStmtRequestList[queryName] = entry

	return nil
}

func (r *Queries) NamedStmtRequest(queryName queryName) (*sqlx.NamedStmt, error) {
	entry, exists := r.namedStmtRequestList[queryName]
	if !exists {
		return nil, fmt.Errorf("%s NamedStmt query not found", queryName)
	}

	return entry.NamedStmt, nil
}

func (r *Queries) AddStmtRequest(queryName queryName, query string) (err error) {
	entry := r.stmtRequestList[queryName]
	entry.Stmt, err = sqlx.Preparex(r.db.Connection(), query)
	if err != nil {
		return fmt.Errorf("%s failed to prepare query, %w", queryName, err)
	}

	entry.Request = query
	r.stmtRequestList[queryName] = entry

	return nil
}

func (r *Queries) StmtRequest(queryName queryName) (*sqlx.Stmt, error) {
	entry, exists := r.stmtRequestList[queryName]
	if !exists {
		return nil, fmt.Errorf("%s StmtRequest query not found", queryName)
	}

	return entry.Stmt, nil
}

func (dbr *Queries) CloseRequests() error {
	for _, request := range dbr.namedStmtRequestList {
		if err := request.NamedStmt.Close(); err != nil {
			return err
		}
	}

	for _, request := range dbr.stmtRequestList {
		if err := request.Stmt.Close(); err != nil {
			return err
		}
	}

	return nil
}
