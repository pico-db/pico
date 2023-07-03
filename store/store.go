package store

type Store interface {
	// Start a transaction.
	//
	// Set isWrite to false for read operations
	// since it avoids Badger's overhead for write operations.
	Start(isWrite bool) (Transaction, error)

	// Close the database. Can be called multiple times
	Close() error
}

type Transaction interface {
	// Set a value to be associated to a key
	Set(key, value []byte) error

	// Get the value based on the key
	Get(key []byte) ([]byte, error)

	// Delete the value based on the key
	Delete(key []byte) error

	// Returns a cursor for iterating multiple values.
	//
	// Set isForward to true if needed to iterate from smallest key to largest
	Cursor(isForward bool) (Cursor, error)

	// Commit the trasaction.
	// This is crucial for changes to be made into the database
	Commit() error

	// Rollback, cancel the transaction
	Rollback() error
}

type Cursor interface {
	// Move the iterator to the provided key.
	// Returns the next smallest key if travelled in forward.
	// If reverse, returns the next largest key
	Seek(key []byte) error

	// Advance the iterator by one.
	// If forward, it goes to the larger key and reverse does the reverse of that
	Next()

	// Returns true when iteration is done
	IsDone() bool

	// Returns the item at the current iteration
	Item() (Item, error)

	// Crucial to call this after finished with the iteration
	Close() error
}

type Item struct {
	Key   []byte
	Value []byte
}
