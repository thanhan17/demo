package grpc

import (
	"context"
	"database/sql"
	"github.com/thanhan17/demo/model/v1"
	"google.golang.org/grpc"
	"net"
)

func RunServer(ctx context.Context, v1API v1.ModelServiceServer, port string) error {
	listen, err = net.listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	//register service
	server := grpc.NewServer()
	v1.RegisterToDoServiceServer(server, v1API)

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			// sig is a ^C, handle it
			log.Println("shutting down gRPC server...")
			server.GracefulStop()
			<-ctx.Done()
		}
	}()

	log.Println("starting gRPC server...")
	return server.Serve(listen)
}
