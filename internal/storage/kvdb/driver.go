package kvdb

// Driver represents a KeyValueStore database driver.
//
// The Driver implementation pattern follows the same registration
// pattern as database/sql. Namely, implementations of the KeyValueStore
// interface should implement the Driver pattern to obtain a handle to
// a KeyValueStore.
type Driver interface {
	Open(uri string) (KeyValueStore, error)
}
