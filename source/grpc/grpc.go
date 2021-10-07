package main

import (
	"context"
	"net"
	"os"
	"os/signal"

	v1 "github.com/thanhan17/demo/source/grpc/model/v1"
	"google.golang.org/grpc"
)

func runServer(ctx context.Context, v1API v1.UserServiceServer, port string) error {
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	//register service
	server := grpc.NewServer()
	v1.RegisterUserServiceServer(server, v1API)

	// graceful shutdown
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			// sig is a ^C, handle it
			log.Info("shutting down gRPC server...")
			server.GracefulStop()
			<-ctx.Done()
		}
	}()

	log.Info("Starting gRPC server...")
	return server.Serve(listen)
}
