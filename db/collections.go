package db

import (
	"encoding/json"
	"fmt"

	"github.com/pico-db/pico/internal/utils"
	"github.com/pico-db/pico/store"
)

type DB struct {
	s store.Store
}

type collectionMetadata struct {
	Size int `json:"size"`
}

// Create a collection in the database
func (db *DB) CreateCollection(name string) error {
	return db.createCollection(name)
}

// Remove a collection from the database, removing all documents
func (db *DB) DropCollection(name string) error {
	return db.dropCollection(name)
}

func (db *DB) dropCollection(name string) error {
	return db.tranact(true, func(tx store.Transaction) error {
		err := tx.Delete(utils.ToBytes(db.getCollectionName(name)))
		if err != nil {
			return err
		}
		return nil
	})
}

// Perform actions inside a transaction
func (db *DB) Transact(isWrite bool, do TransactionFunc) error {
	return db.tranact(isWrite, do)
}

func (db *DB) tranact(isWrite bool, do TransactionFunc) error {
	tx, err := db.s.Start(isWrite)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = do(tx)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (db *DB) createCollection(name string) error {
	return db.tranact(true, func(tx store.Transaction) error {
		yes, err := db.hasCollection(name, tx)
		if err != nil {
			return err
		}
		if yes {
			return ErrCollectionExists
		}
		meta := collectionMetadata{
			Size: 0,
		}
		err = db.saveCollectionMetadata(name, &meta, tx)
		if err != nil {
			return err
		}
		return nil
	})
}

func (db *DB) saveCollectionMetadata(col string, meta *collectionMetadata, tx store.Transaction) error {
	r, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return tx.Set(utils.ToBytes(db.getCollectionName(col)), r)
}

func (db *DB) hasCollection(name string, tx store.Transaction) (bool, error) {
	v, err := tx.Get(utils.ToBytes(db.getCollectionName(name)))
	if err != nil {
		return false, err
	}
	exists := utils.NotNil(v)
	return exists, nil
}

func (db *DB) getCollectionName(name string) string {
	return fmt.Sprintf("%v%v", db.getCollectionPrefix(), name)
}

func (db *DB) getCollectionPrefix() string {
	return "coll:"
}
