package db

import (
	"errors"

	"github.com/pico-db/pico/store"
)

var (
	ErrCollectionExists = errors.New("collection already exists")
	ErrIdNotFound       = errors.New("field not found")
	ErrInvalidId        = errors.New("invalid id type")
)

const (
	ObjectIdField  = "_id"
	ExpiresAtField = "_expiresAt"
)

type TransactionFunc = func(tx store.Transaction) error