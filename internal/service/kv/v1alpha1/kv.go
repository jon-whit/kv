package v1alpha1

import (
	"context"
	"errors"
	"log"

	kvv1alpha1 "github.com/jon-whit/kv/internal/proto/kv/v1alpha1"
	"github.com/jon-whit/kv/internal/storage/kvdb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	errKeyNotFoundMessageFmt = "key '%s' not found"
	errInternalServerMessage = "an internal server error has occurred"
)

// kvService implements the KVService service interface.
type kvService struct {
	kvdb kvdb.KeyValueStore
	kvv1alpha1.UnimplementedKVServiceServer
}

var _ kvv1alpha1.KVServiceServer = (*kvService)(nil)

func NewKVService(kvdb kvdb.KeyValueStore) kvv1alpha1.KVServiceServer {
	return &kvService{
		kvdb: kvdb,
	}
}

// Delete implements kvv1alpha1.KVServiceServer.
func (k *kvService) Delete(
	ctx context.Context,
	req *kvv1alpha1.DeleteRequest,
) (*kvv1alpha1.DeleteResponse, error) {
	_, err := k.kvdb.DeleteItem(ctx, kvdb.DeleteItemRequest{
		Key: req.Key,
	})
	if err != nil {
		if errors.Is(err, kvdb.ErrKeyNotFound) {
			return nil, status.Errorf(codes.NotFound, errKeyNotFoundMessageFmt, req.GetKey())
		}
	}

	return &kvv1alpha1.DeleteResponse{}, nil
}

// Get implements kvv1alpha1.KVServiceServer.
func (k *kvService) Get(
	ctx context.Context,
	req *kvv1alpha1.GetRequest,
) (*kvv1alpha1.GetResponse, error) {
	getItemResp, err := k.kvdb.GetItem(ctx, kvdb.GetItemRequest{
		Key: req.GetKey(),
	})
	if err != nil {
		if errors.Is(err, kvdb.ErrKeyNotFound) {
			return nil, status.Errorf(codes.NotFound, errKeyNotFoundMessageFmt, req.GetKey())
		}
	}

	var val structpb.Value
	err = proto.Unmarshal(getItemResp.Value, &val)
	if err != nil {
		log.Printf("failed to proto.Unmarshal value from storage layer: %v\n", err)
		return nil, status.Errorf(codes.Internal, errInternalServerMessage)
	}

	return &kvv1alpha1.GetResponse{
		Key:   getItemResp.Key,
		Value: &val,
	}, nil
}

// Set implements kvv1alpha1.KVServiceServer.
func (k *kvService) Set(
	ctx context.Context,
	req *kvv1alpha1.SetRequest,
) (*kvv1alpha1.SetResponse, error) {
	byteVal, err := proto.Marshal(req.GetValue())
	if err != nil {
		log.Printf("failed to proto.Marshal value: %v\n", err)
		return nil, status.Errorf(codes.Internal, errInternalServerMessage)
	}

	_, err = k.kvdb.PutItem(ctx, kvdb.PutItemRequest{
		Key:   req.GetKey(),
		Value: byteVal,
	})
	if err != nil {
		log.Printf("kvdb.PutItem failed with error: %v\n", err)
		return nil, status.Errorf(codes.Internal, errInternalServerMessage)
	}

	return &kvv1alpha1.SetResponse{}, nil
}
