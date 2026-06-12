package repository

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// PGErrorClassification identifies whether a PostgreSQL error can be retried
type PGErrorClassification int

const (
	// NonRetriable means the failed operation should not be retried
	NonRetriable PGErrorClassification = iota

	// Retriable means the failed operation may be retried
	Retriable
)

// PostgresErrorClassifier classifies PostgreSQL errors by their retryability
type PostgresErrorClassifier struct{}

// NewPostgresErrorClassifier creates a PostgreSQL error classifier
func NewPostgresErrorClassifier() *PostgresErrorClassifier {
	return &PostgresErrorClassifier{}
}

// Classify returns the retry classification for err
func (c *PostgresErrorClassifier) Classify(err error) PGErrorClassification {
	if err == nil {
		return NonRetriable
	}

	// Check whether err wraps a PostgreSQL error
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return СlassifyPgError(pgErr)
	}

	// Treat unknown errors as non-retryable by default
	return NonRetriable
}

// СlassifyPgError returns the retry classification for a PostgreSQL error
func СlassifyPgError(pgErr *pgconn.PgError) PGErrorClassification {
	// PostgreSQL error codes: https://www.postgresql.org/docs/current/errcodes-appendix.html

	switch pgErr.Code {
	// Class 08 - connection exceptions
	case pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure:
		return Retriable

	// Class 40 - transaction rollback errors
	case pgerrcode.TransactionRollback, // 40000
		pgerrcode.SerializationFailure, // 40001
		pgerrcode.DeadlockDetected:     // 40P01
		return Retriable

	// Class 57 - operator intervention errors
	case pgerrcode.CannotConnectNow: // 57P03
		return Retriable
	}

	// Explicitly mark known client-side and schema errors as non-retryable
	switch pgErr.Code {
	// Class 22 - data exceptions
	case pgerrcode.DataException,
		pgerrcode.NullValueNotAllowedDataException:
		return NonRetriable

	// Class 23 - integrity constraint violations
	case pgerrcode.IntegrityConstraintViolation,
		pgerrcode.RestrictViolation,
		pgerrcode.NotNullViolation,
		pgerrcode.ForeignKeyViolation,
		pgerrcode.UniqueViolation,
		pgerrcode.CheckViolation:
		return NonRetriable

	// Class 42 - syntax and access rule violations
	case pgerrcode.SyntaxErrorOrAccessRuleViolation,
		pgerrcode.SyntaxError,
		pgerrcode.UndefinedColumn,
		pgerrcode.UndefinedTable,
		pgerrcode.UndefinedFunction:
		return NonRetriable
	}

	// Treat unknown PostgreSQL errors as non-retryable by default
	return NonRetriable
}
