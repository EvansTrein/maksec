package postgres

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"maksec/internal/config"
)

type Option func(d *Database)

type Database struct {
	dsn  string
	conn *sqlx.DB
	ctx  context.Context
}

func (d *Database) Connection() *sqlx.DB {
	return d.conn
}

func (d *Database) SqlConnection() *sql.DB {
	return d.conn.DB
}

func (d *Database) Ping() error {
	return d.conn.PingContext(d.ctx)
}

func (d *Database) Context() context.Context {
	return d.ctx
}

func (d *Database) Close() error {
	if d.conn == nil {
		return nil
	}
	return d.conn.Close()
}

func New(dsn string, opts ...Option) (*Database, error) {
	d := &Database{
		dsn: dsn,
		ctx: config.DefaultCtx(),
	}

	for _, o := range opts {
		o(d)
	}

	conn, err := sqlx.Open("pgx", d.dsn)
	if err != nil {
		return nil, err
	}

	conn.SetMaxIdleConns(3)
	conn.SetConnMaxIdleTime(20 * time.Second)
	conn.SetMaxOpenConns(10)

	d.conn = conn

	if err := d.conn.PingContext(d.ctx); err != nil {
		return nil, err
	}

	return d, nil
}

func WithContext(ctx context.Context) Option {
	return func(p *Database) {
		p.ctx = ctx
	}
}
