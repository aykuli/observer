package postgres

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type postgresError struct {
	name string
	Err  error
}

func (dbErr *postgresError) Error() string {
	return "[" + dbErr.name + "] " + dbErr.Err.Error()
}

func newDBError(err error) error {
	var pgErr *pgconn.PgError
	name := "Postgres"
	if errors.As(err, &pgErr) {
		if pgerrcode.IsConnectionException(pgErr.Code) {
			name = "Postgres. Connection Exception"
		}

		if pgerrcode.IsDataException(pgErr.Code) {
			name = "Postgres. Data Exception"
		}

	}

	return &postgresError{
		name: name,
		Err:  err,
	}
}
