package pgstore

import "errors"

var (
	ErrInvalidCredentials = errors.New("pgstore: invalid credentials")
	ErrNoRecord           = errors.New("pgstore: no matching record found")
	ErrDuplicateEmail     = errors.New("pgstore: invalid duplicate email")
)
