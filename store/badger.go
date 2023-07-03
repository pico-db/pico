package store

import (
	"errors"

	"github.com/dgraph-io/badger/v3"
)

var (
	ErrKeyNotFound     = errors.New("key not found")
	ErrCursorItemEmpty = errors.New("empty item found")
)

// Badger implementation of the Store interface
type badgerStore struct {
	db *badger.DB
}

type badgerTransaction struct {
	tx *badger.Txn
}

type badgerCursor struct {
	it *badger.Iterator
}

// Initialize or open existing Badger KV database
func Open(dir string) (Store, error) {
	return OpenWithOptions(
		badger.DefaultOptions(dir),
	)
}

// Initialize or open existing Badger KV database with options
func OpenWithOptions(opts badger.Options) (Store, error) {
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	s := &badgerStore{
		db: db,
	}
	return s, nil
}

func (s *badgerStore) Close() error {
	return s.db.Close()
}

func (s *badgerStore) Start(isWrite bool) (Transaction, error) {
	t := s.db.NewTransaction(isWrite)
	return &badgerTransaction{
		tx: t,
	}, nil
}

func (t *badgerTransaction) Set(key, value []byte) error {
	return t.tx.Set(key, value)
}

func (t *badgerTransaction) Get(key []byte) ([]byte, error) {
	it, err := t.tx.Get(key)
	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, ErrKeyNotFound
	}
	if err != nil {
		return nil, err
	}
	return toItem(it)
}

func toItem(it *badger.Item) ([]byte, error) {
	var v []byte
	err := it.Value(func(val []byte) error {
		v = val
		return nil
	})
	return v, err
}

func (t *badgerTransaction) Delete(key []byte) error {
	return t.tx.Delete(key)
}

func (t *badgerTransaction) Commit() error {
	return t.tx.Commit()
}

func (t *badgerTransaction) Rollback() error {
	t.tx.Discard()
	return nil
}

func (t *badgerTransaction) Cursor(isForward bool) (Cursor, error) {
	opts := badger.DefaultIteratorOptions
	opts.Reverse = !isForward
	return &badgerCursor{
		it: t.tx.NewIterator(opts),
	}, nil
}

func (t *badgerCursor) Seek(key []byte) error {
	t.it.Seek(key)
	return nil
}

func (t *badgerCursor) Next() {
	t.it.Next()
}

func (t *badgerCursor) IsDone() bool {
	return !t.it.Valid()
}

func (t *badgerCursor) Item() (Item, error) {
	it := t.it.Item()
	if it == nil {
		return Item{}, ErrCursorItemEmpty
	}
	v, err := toItem(it)
	return Item{
		Key:   it.Key(),
		Value: v,
	}, err
}

func (c *badgerCursor) Close() error {
	c.it.Close()
	return nil
}
