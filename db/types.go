package db

import (
	"errors"

	"github.com/pico-db/pico/store"
)

var (
	ErrCollectionExists = errors.New("collection already exists")
	ErrIdNotFound       = errors.New("field not found")
	ErrInvalidId        = errors.New("invalid id type")
	ErrUnmarshallable   = errors.New("provided object is not a map or a struct")
)

const (
	ObjectIdField  = "_id"
	ExpiresAtField = "_expiresAt"
)

type TransactionFunc = func(tx store.Transaction) error
