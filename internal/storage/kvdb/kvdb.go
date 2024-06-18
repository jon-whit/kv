package kvdb

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
)

var (
	driversMu sync.RWMutex
	drivers   = make(map[string]Driver)

	ErrKeyNotFound         = errors.New("no item with the provided key exists")
	ErrTransactionConflict = errors.New("a transaction conflict has occurred")
)

// KeyValueStore defines an interface for managing items in Key/Value database.
type KeyValueStore interface {

	// PutItem idempotently puts the provided item in the request. If the key
	// being put is concurrently being modified then an ErrTransactionConflict
	// should be returned. Otherwise the item should be upserted to the underlying
	// storage.
	PutItem(context.Context, PutItemRequest) (PutItemResponse, error)

	// GetItem fetches the item of the provided key. If the item of that key does
	// not exist it should return ErrKeyNotFound.
	GetItem(context.Context, GetItemRequest) (GetItemResponse, error)

	// DeleteItem idempotently removes an item of the provided key. If the key is actively
	// being upsert or mutated in some way an ErrTransactionConflict is returned.
	DeleteItem(context.Context, DeleteItemRequest) (DeleteItemResponse, error)

	// Close closes the KeyValueStore and should flush any unwritten data to the underlyling storage
	// layer.
	Close(context.Context) error
}

type PutItemRequest struct {
	Key   string
	Value []byte
}

type PutItemResponse struct{}

type GetItemRequest struct {
	Key string
}

type GetItemResponse struct {
	Key   string
	Value []byte
}

type DeleteItemRequest struct {
	Key string
}

type DeleteItemResponse struct{}

// Register registers the provided driver under the given name.
// If a driver with the provided name is already registered this
// function will panic.
func Register(name string, driver Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()

	_, ok := drivers[name]
	if ok {
		panic(fmt.Sprintf("duplicate registration of kv driver '%s'", name))
	}

	drivers[name] = driver
}

// Drivers returns a sorted list of the names of the registered drivers.
func Drivers() []string {
	driversMu.RLock()
	defer driversMu.RUnlock()
	list := make([]string, 0, len(drivers))
	for name := range drivers {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

// Open opens the KeyValueStore by delegating to the provided driver.
//
// Calling code should ensure that the driver is registered first by
// calling Register or by importing the approapriate driver. For example,
//
// import (_ "github.com/jon-whit/kv/internal/storage/kvdb/badgerdb")
func Open(driverName, dataSourceName string) (KeyValueStore, error) {
	driversMu.RLock()
	driveri, ok := drivers[driverName]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("kvdb: unknown storage driver %q", driverName)
	}

	return driveri.Open(dataSourceName)
}
