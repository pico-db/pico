package db

import (
	"errors"

	"github.com/pico-db/pico/internal/umap"
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

// Represents a document in a collection
type Document struct {
	fields map[string]interface{}
}

// Returns the document ID.
//   - ErrIdNotFound if the ID is not found (key is "_id").
//   - ErrInvalidId if the ID has an invalid type (must be string)
func (d *Document) ObjectId() (string, error) {
	return d.objectId()
}

// Returns the value associated with a key inside the document.
// It returns nil if the key is not found, or the value is nil
func (d *Document) Get(key string) interface{} {
	return d.get(key)
}

func (d *Document) objectId() (string, error) {
	id := d.get(ObjectIdField)
	if id == nil {
		return "", ErrIdNotFound
	}
	s, yes := id.(string)
	if !yes {
		return "", ErrInvalidId
	}
	return s, nil
}

func (d *Document) get(k string) interface{} {
	_, val, _ := umap.Lookup(d.fields, k)
	return val
}
