package v1

import (
	"log"
	"database/sql"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/thanhan17/model/v1"
)

const (
	apiVersion = "v1"
)

type modelServiceServer struct {
	db *sql.DB
}

type NewModelService(db *sql.DB) v1.modelServiceServer{
	return &modelServiceServer{db: db}
}

func (s *modelServiceServer) Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error) {
	log.Printf("Receive message body client: %s", req.Api, req.User)
	return &CreateResponse{Api: apiVersion, Id: "001"}, nil
}

func (s *modelServiceServer) Read(ctx context.Context, req *ReadRequest) (*ReadResponse, error) {
	log.Printf("Receive message body client: %s", req.Api, req.User)
	return &ReadResponse{Api: apiVersion, &User{Id: req.Id, Name: "An", Phone: "0905"}}, nil
}

func (s *modelServiceServer) Update(ctx context.Context, req *UpdateRequest) (*UpdateResponse, error) {
	log.Printf("Receive message body client: %s", req.Api, req.User)
	return &UpdateResponse{Api: apiVersion, Updated: 1}, nil
}

func (s *modelServiceServer) Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	log.Printf("Receive message body client: %s", req.Api, req.User)
	return &DeleteResponse{Api: apiVersion, Deleted: req.Id}, nil
}

func (s *modelServiceServer) ReadAll(ctx context.Context, req *ReadAllRequest) (*ReadAllResponse, error) {
	log.Printf("Receive message body client: %s", req.Api, req.User)
	return &ReadAllResponse{Api: apiVersion, Body: "Read all From the Server!"}, nil
}
