package migrate

import (
	"context"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Up(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	path := filepath.Join(migrationsDir, "001_transactions.up.sql")
	sql, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, string(sql))
	return err
}
