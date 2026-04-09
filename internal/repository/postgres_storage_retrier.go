package repository

import (
	"context"
	"database/sql"
	"fmt"
)

func (s *PostgresStorage) withRetryTx(ctx context.Context, executeFunction func(*sql.Tx) error) error {
	const maxAttempts = 3
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			if s.isRetriablePostgresError(err) && attempt < maxAttempts {
				lastErr = err
				continue
			}
			return fmt.Errorf("begin tx: %w", err)
		}

		err = executeFunction(tx)
		if err != nil {
			_ = tx.Rollback()

			if s.isRetriablePostgresError(err) && attempt < maxAttempts {
				lastErr = err
				continue
			}

			return err
		}

		err = tx.Commit()
		if err != nil {
			_ = tx.Rollback()

			if s.isRetriablePostgresError(err) && attempt < maxAttempts {
				lastErr = err
				continue
			}

			return fmt.Errorf("commit tx: %w", err)
		}

		return nil
	}

	return fmt.Errorf("transaction failed after retries: %w", lastErr)
}
