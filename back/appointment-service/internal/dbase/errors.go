package dbase

import (
	"database/sql"
	"errors"
	"fmt"
	common "scheduler/appointment-service/internal"
)

func DbError(e error) error {
	if e == nil {
		return nil
	}
	if errors.Is(e, sql.ErrNoRows) {
		e = common.ErrNotFound
	}
	return fmt.Errorf("db error: %w", e)
}
