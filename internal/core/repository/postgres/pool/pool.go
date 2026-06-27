package core_postgres_pool

import (
	"context"
	core_config "discord/internal/core/config"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKeyType struct{}

var txKey = txKeyType{}

type Pool interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	Close()

	CtxWithTransaction(
		ctx context.Context,
		fn func(ctx context.Context) error,
	) error
	GetTx(
		ctx context.Context,
	) pgx.Tx
	OpTimeout() time.Duration
}

type ConnectionPool struct {
	*pgxpool.Pool
	opTimeout time.Duration
}

func NewConnectionPool(ctx context.Context, cfg core_config.Config_DB) (*ConnectionPool, error) {
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name,
	)
	pgxconfig, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("parse pgx config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxconfig)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pgx pool ping: %w", err)
	}

	return &ConnectionPool{
		Pool:      pool,
		opTimeout: cfg.Timeout,
	}, nil
}

func (p *ConnectionPool) OpTimeout() time.Duration {
	return p.opTimeout
}

func (p *ConnectionPool) GetTx(
	ctx context.Context,
) pgx.Tx {
	if tx, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return tx
	}
	return nil
}
func (p *ConnectionPool) CtxWithTransaction(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	if tx := p.GetTx(ctx); tx != nil {
		return fn(ctx)
	}
	tx, err := p.Begin(ctx)
	if err != nil {
		return fmt.Errorf("create context: %w", err)
	}
	defer tx.Rollback(ctx)

	ctx_tx := context.WithValue(ctx, txKey, tx)
	if err := fn(ctx_tx); err != nil {
		return fmt.Errorf("execute transaction fn: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
	//ctx := context.WithValue(r.Context(), "log", l)
}
