package v1

import (
	"context"
	"database/sql"
	"log"
)

const (
	apiVersion = "v1"
)

type userServiceServer struct {
	db *sql.DB
}

func NewUserServiceServer(db *sql.DB) UserServiceServer {
	return &userServiceServer{db: db}
}

func (s *userServiceServer) Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error) {
	log.Printf("Create:\tAPI: %s\tUser: %+v", req.Api, req.User)
	return &CreateResponse{Api: apiVersion, Id: 1}, nil
}

func (s *userServiceServer) Read(ctx context.Context, req *ReadRequest) (*ReadResponse, error) {
	log.Printf("Read\tAPI: %s\tId: %d", req.Api, req.Id)
	return &ReadResponse{Api: apiVersion, User: &User{Id: req.Id, Name: "An", Phone: "0905"}}, nil
}

func (s *userServiceServer) Update(ctx context.Context, req *UpdateRequest) (*UpdateResponse, error) {
	log.Printf("Update:\tAPI: %s\tUser: %+v", req.Api, req.User)
	return &UpdateResponse{Api: apiVersion, Updated: 1}, nil
}

func (s *userServiceServer) Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	log.Printf("Delete:\tAPI: %s\tId: %d", req.Api, req.Id)
	return &DeleteResponse{Api: apiVersion, Deleted: req.Id}, nil
}

func (s *userServiceServer) ReadAll(ctx context.Context, req *ReadAllRequest) (*ReadAllResponse, error) {
	log.Printf("Read all:\tAPI: %s", req.Api)
	users := []*User{}
	users = append(users, &User{Id: 1, Name: "An", Phone: "0905"})
	return &ReadAllResponse{Api: apiVersion, Users: users}, nil
}
