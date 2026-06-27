package core_postgres_pool

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	migrate_pg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/stdlib"
)

func RunMigrations(pool *ConnectionPool) error {
	db := stdlib.OpenDBFromPool(pool.Pool)
	defer db.Close()

	driver, err := migrate_pg.WithInstance(db, &migrate_pg.Config{})
	if err != nil {
		return fmt.Errorf("migrate postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
