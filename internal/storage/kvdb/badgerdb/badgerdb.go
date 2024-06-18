package badgerdb

import (
	"context"
	"errors"

	badgerv4 "github.com/dgraph-io/badger/v4"
	"github.com/jon-whit/kv/internal/storage/kvdb"
)

func init() {
	kvdb.Register("badgerdb", &badgerKVDriver{})
}

type badgerKVDriver struct{}

type badgerKVStore struct {
	*badgerv4.DB
}

var _ kvdb.KeyValueStore = (*badgerKVStore)(nil)

// Open implements kvdb.Driver.
func (b *badgerKVDriver) Open(uri string) (kvdb.KeyValueStore, error) {
	opts := badgerv4.DefaultOptions(uri)
	opts.Logger = nil // disable default debug logger
	badgerDB, err := badgerv4.Open(opts)
	if err != nil {
		return nil, err
	}

	return &badgerKVStore{
		DB: badgerDB,
	}, nil
}

// DeleteItem implements kvdb.KeyValueStore.
func (b *badgerKVStore) DeleteItem(ctx context.Context, req kvdb.DeleteItemRequest) (kvdb.DeleteItemResponse, error) {
	key := []byte(req.Key)

	txn := b.DB.NewTransaction(true)
	defer txn.Discard()

	_, err := txn.Get(key)
	if err != nil {
		if errors.Is(err, badgerv4.ErrKeyNotFound) {
			return kvdb.DeleteItemResponse{}, kvdb.ErrKeyNotFound
		}

		return kvdb.DeleteItemResponse{}, err
	}

	if err = txn.Delete(key); err != nil {
		return kvdb.DeleteItemResponse{}, err
	}

	if err = txn.Commit(); err != nil {
		if errors.Is(err, badgerv4.ErrConflict) {
			return kvdb.DeleteItemResponse{}, kvdb.ErrTransactionConflict
		}
	}

	return kvdb.DeleteItemResponse{}, nil
}

// GetItem implements kvdb.KeyValueStore.
func (b *badgerKVStore) GetItem(ctx context.Context, req kvdb.GetItemRequest) (kvdb.GetItemResponse, error) {
	var itemValue []byte
	err := b.DB.View(func(txn *badgerv4.Txn) error {
		item, err := txn.Get([]byte(req.Key))
		if err != nil {
			return err
		}

		itemValue, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, badgerv4.ErrKeyNotFound) {
			return kvdb.GetItemResponse{}, kvdb.ErrKeyNotFound
		}
	}

	return kvdb.GetItemResponse{
		Key:   req.Key,
		Value: itemValue,
	}, nil
}

// PutItem implements kvdb.KeyValueStore.
func (b *badgerKVStore) PutItem(
	ctx context.Context,
	req kvdb.PutItemRequest,
) (kvdb.PutItemResponse, error) {
	key := []byte(req.Key)

	err := b.DB.Update(func(txn *badgerv4.Txn) error {
		err := txn.Set(key, req.Value)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, badgerv4.ErrConflict) {
			return kvdb.PutItemResponse{}, kvdb.ErrTransactionConflict
		}

		return kvdb.PutItemResponse{}, err
	}

	return kvdb.PutItemResponse{}, nil
}

// Close implements kvdb.KeyValueStore.
// Subtle: this method shadows the method (*DB).Close of badgerKVStore.DB.
func (b *badgerKVStore) Close(context.Context) error {
	return b.DB.Close()
}
