package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	"go-url-shortener/internal/interceptors"
	"go-url-shortener/internal/service/protoservice"
	"go-url-shortener/internal/service/shortifier"
	"go-url-shortener/proto"

	"google.golang.org/grpc"
)

func GRPCListenAddr(port int) string {
	return ":" + strconv.Itoa(port)
}

func NewGRPCServer(sf *shortifier.Shortifier) *grpc.Server {
	s := grpc.NewServer(grpc.UnaryInterceptor(interceptors.UnaryInterceptor))
	proto.RegisterShortenerServiceServer(s, protoservice.NewShortenerServiceServer(sf))
	return s
}

func ServeGRPC(ctx context.Context, sf *shortifier.Shortifier, listenAddr string) error {
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("grpc listen %s: %w", listenAddr, err)
	}

	return ServeGRPCOnListener(ctx, lis, sf)
}

func ServeGRPCOnListener(ctx context.Context, lis net.Listener, sf *shortifier.Shortifier) error {
	defer func() { _ = lis.Close() }()

	srv := NewGRPCServer(sf)

	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()

	err := srv.Serve(lis)
	if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return fmt.Errorf("grpc serve: %w", err)
	}

	return nil
}
