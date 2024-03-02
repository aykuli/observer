package errors

import (
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	errConflict = errors.New("введены неправильные данные")
	errNoRow    = errors.New("нет таких данных")
)

type StorageError struct {
	method string
	Err    error
}

func (se *StorageError) Error() string {
	return fmt.Sprintf("[STORAGE] %v\n|-- %v\n", se.Err, se.method)
}

func NewStorageError(method string, err error) error {
	return &StorageError{method: method, Err: err}
}

type DBError struct {
	Err error
}

func (re *DBError) Error() string {
	return re.Err.Error()
}

func NewDBError(err error) error {
	respErr := err
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		respErr = errConflict
	}

	if errors.Is(err, pgx.ErrNoRows) {
		respErr = errNoRow
	}

	return &DBError{Err: respErr}
}
