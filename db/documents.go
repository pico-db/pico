package db

import (
	"fmt"
	"time"

	"github.com/pico-db/pico/internal/umap"
	"github.com/pico-db/pico/internal/utils"
	uuid "github.com/satori/go.uuid"
)

// Represents a document in a collection
type Document struct {
	fields map[string]interface{}
}

// Create a new empty document.
// User must add _id and _expiresAt themselves.
//
// It will be validated when being used
func NewDocument() *Document {
	return &Document{
		fields: make(map[string]interface{}),
	}
}

// Create a new document initialized from
// an object.
//
// Please use maps, or else it will return an error
func NewDocumentFrom(from interface{}) (*Document, error) {
	return newDocumentFrom(from)
}

func newDocumentFrom(from interface{}) (*Document, error) {
	doc, isDoc := from.(*Document)
	if isDoc {
		return doc, nil
	}
	normalized, err := utils.Normalize(from)
	if err != nil {
		return nil, err
	}
	mapped, isMap := normalized.(map[string]interface{})
	if !isMap {
		return nil, ErrUnmarshallable
	}
	return &Document{
		fields: mapped,
	}, nil
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

// Check for if the document has valid _id and _expiresAt
func (d *Document) IsValid() error {
	if !d.isValidObjectId() {
		return fmt.Errorf("invalid _id")
	}
	if d.has(ExpiresAtField) && d.expiresAt() == nil {
		return fmt.Errorf("invalid _expiresAt: %s", d.get(ExpiresAtField))
	}
	return nil
}

// Returns the time at which this document is expired
func (d *Document) ExpiresAt() *time.Time {
	return d.expiresAt()
}

// Set the expiration date of this document
func (d *Document) SetExpiresAt(exp time.Time) error {
	return d.setExpiresAt(exp)
}

// Encodes the document into bytes array using MessagePack encoding
func (d *Document) Encode() ([]byte, error) {
	return d.encode()
}

// Decodes the byte arrays into the document.
//
// Will resets all of the document's existing fields and values
func (d *Document) Decode(data []byte) error {
	return d.decode(data)
}

// Takes in an interface and marshal it into the current document
//
// Will resets all of the document's existing fields and values
func (d *Document) Marshal(from interface{}) error {
	return d.marshal(from)
}

// Unpacks the document values into a struct
func (d *Document) Unmarshal(to interface{}) error {
	return d.unmarshal(to)
}

func (d *Document) marshal(from interface{}) error {
	doc, isDoc := from.(*Document)
	if isDoc {
		d.fields = doc.fields
		return nil
	}
	normalized, err := utils.Normalize(from)
	if err != nil {
		return err
	}
	mapped, isMap := normalized.(map[string]interface{})
	if !isMap {
		return ErrUnmarshallable
	}
	d.fields = mapped
	return nil
}

func (d *Document) unmarshal(to interface{}) error {
	doc, isDoc := to.(*Document)
	if isDoc {
		doc.fields = d.fields
		return nil
	}
	return umap.Unmarshal(d.fields, to)
}

func (d *Document) decode(data []byte) error {
	d.fields = make(map[string]interface{})
	return umap.Decode(data, &d.fields)
}

func (d *Document) encode() ([]byte, error) {
	return umap.Encode(d.fields)
}

func (d *Document) isValidObjectId() bool {
	docid, err := d.ObjectId()
	if err != nil {
		return false
	}
	_, err = uuid.FromString(docid)
	return err == nil
}

func (d *Document) expiresAt() *time.Time {
	expiry, ok := d.get(ExpiresAtField).(time.Time)
	if !ok {
		return nil
	}
	return &expiry
}

func (d *Document) setExpiresAt(exp time.Time) error {
	return d.upsert(ExpiresAtField, exp)
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
