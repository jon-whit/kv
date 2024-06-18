package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	kvv1alpha1 "github.com/jon-whit/kv/internal/proto/kv/v1alpha1"
	kvv1alpha1svc "github.com/jon-whit/kv/internal/service/kv/v1alpha1"
	"github.com/jon-whit/kv/internal/storage/kvdb"
	_ "github.com/jon-whit/kv/internal/storage/kvdb/badgerdb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	MaxRequestBytes = 1.5 * 1024 * 1024 // 1.5MiB
)

var addrFlag = flag.String("addr", ":50052", "the address to serve the KV service on")

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kvdb, err := kvdb.Open("badgerdb", "/tmp/badger")
	if err != nil {
		log.Fatalf("failed to initialize underlying kvdb: %v", err)
	}
	defer kvdb.Close(ctx)

	kvsvc := kvv1alpha1svc.NewKVService(kvdb)

	lis, err := net.Listen("tcp", *addrFlag)
	if err != nil {
		log.Fatalf("failed to start tcp listener: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(MaxRequestBytes), // limit max request size
	}

	grpcServer := grpc.NewServer(opts...)
	reflection.Register(grpcServer)
	kvv1alpha1.RegisterKVServiceServer(grpcServer, kvsvc)

	log.Printf("starting KV service on '%s'...\n", *addrFlag)
	if err := grpcServer.Serve(lis); err != nil {
		if !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("failed to start grpc server: %v", err)
		}
	}

	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig
}
