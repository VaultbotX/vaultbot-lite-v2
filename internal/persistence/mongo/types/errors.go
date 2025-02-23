package types

import "errors"

var (
	ErrNoDocuments = errors.New("mongo: no documents in result")
)
