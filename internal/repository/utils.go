package repository

import (
	"errors"
	"strings"

	"github.com/lib/pq"
)

func mapPGErrorToRepoError(err error) error {
	if pgErr, ok := errors.AsType[*pq.Error](err); ok {
		if pgErr.Code == "23503" { // FK constraint violation
			if strings.Contains(pgErr.Constraint, "user_id_fkey") {
				return ErrUserNotFound
			}
			if strings.Contains(pgErr.Constraint, "item_id_fkey") {
				return ErrItemNotFound
			}
		}
	}
	return err
}
