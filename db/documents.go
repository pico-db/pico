package db

import (
	"fmt"

	"github.com/pico-db/pico/internal/umap"
	"github.com/pico-db/pico/internal/utils"
)

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

// Check if the provided key exists inside the document
func (d *Document) Has(key string) bool {
	return d.has(key)
}

// Set a value to a key inside the document.
// By default, it overrides the existing value associated to the key
// and also add a new key if key does not exist
func (d *Document) Set(key string, val interface{}) error {
	return d.upsert(key, val)
}

// Update the fields inside the document.
// Set upsert to true if new fields are meant to be inserted into the document.
//
// Upsert true is equal to calling Set on all fields
func (d *Document) Update(updates map[string]interface{}, upsert bool) error {
	for k, v := range updates {
		var err error
		if upsert {
			err = d.upsert(k, v)
		} else {
			err = d.update(k, v)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Returns the document as a map
func (d *Document) Map() map[string]interface{} {
	return umap.Copy(d.fields)
}

// Returns a list of sorted keys inside the document.
func (d *Document) Fields(withSubFields bool) []string {
	return umap.Keys(d.fields, true, withSubFields)
}



func (d *Document) update(key string, val interface{}) error {
	normal, err := utils.Normalize(val)
	if err != nil {
		return err
	}
	// fails if the key does not exists
	parent, _, lastKeyInProvided := umap.Lookup(d.fields, key, false)
	if parent == nil {
		return fmt.Errorf("key not found: %s", key)
	}
	parent[lastKeyInProvided] = normal
	return nil
}

func (d *Document) upsert(key string, val interface{}) error {
	normal, err := utils.Normalize(val)
	if err != nil {
		return err
	}
	// travel to the nearest parent
	parent, _, lastKeyInProvided := umap.Lookup(d.fields, key, true)
	parent[lastKeyInProvided] = normal
	return nil
}

func (d *Document) has(key string) bool {
	parent, _, _ := umap.Lookup(d.fields, key, false)
	return parent != nil
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
	_, val, _ := umap.Lookup(d.fields, k, false)
	return val
}
