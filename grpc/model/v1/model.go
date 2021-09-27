package v1

import (
	"context"
	"database/sql"
	fmt "fmt"
	"log"

	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

func (s *userServiceServer) connect(ctx context.Context) (*sql.Conn, error) {
	c, err := s.db.Conn(ctx)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to connect to database-> "+err.Error())
	}
	return c, nil
}

func (s *userServiceServer) checkAPI(api string) error {
	if len(api) > 0 {
		if apiVersion != api {
			return status.Errorf(codes.Unimplemented,
				"Unsupported API version: service implements API version '%s', but asked for '%s'", apiVersion, api)
		}
	}
	return nil
}

// Not check exist and validation after insert
func (s *userServiceServer) Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error) {
	log.Printf("Create: " + req.String())
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	//SQL connection pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	//Insert entity user data
	res, err := c.ExecContext(ctx, "INSERT INTO user(`id`, `name`, `phone`) values (?, ?, ?)", 1, req.User.Name, req.User.Phone)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to insert into user -> "+err.Error())
	}

	//Return id user created
	id, err := res.LastInsertId()
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to retrieve id for created user -> "+err.Error())
	}
	log.Printf("%T", id)
	return &CreateResponse{Api: req.Api, Id: id}, nil
}

func (s *userServiceServer) Read(ctx context.Context, req *ReadRequest) (*ReadResponse, error) {
	log.Printf("Read: " + req.String())
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}
	//SQL connection pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	//Query user by id
	row, err := c.QueryContext(ctx, "SELECT * FROM user where id = ?", req.Id)
	if err != nil {
		if err := row.Err(); err != nil {
			return nil, status.Error(codes.Unknown, "Failed to retrieve data from user -> "+err.Error())
		}
		return nil, status.Error(codes.NotFound, "User with id='%d' not found -> "+err.Error())
	}
	defer row.Close()
	//Retrieve data struct user
	var user User
	if row.Next() {
		if err := row.Scan(&user.Id, &user.Name, &user.Phone); err != nil {
			return nil, status.Error(codes.NotFound, "Failed to retrieve field values from user -> "+err.Error())
		}
	}
	if row.Next() {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("Found multiple user rows with id='%d'", req.Id))
	}
	return &ReadResponse{Api: req.Api, User: &user}, nil
}

func (s *userServiceServer) Update(ctx context.Context, req *UpdateRequest) (*UpdateResponse, error) {
	log.Printf("Update:\tAPI: %s\tUser: %+v", req.Api, req.User)
	//SQL connection pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	// update ToDo
	res, err := c.ExecContext(ctx, "UPDATE user SET `name`=?, `phone`=? WHERE `id`=?",
		req.User.Id, req.User.Name, req.User.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to update user -> "+err.Error())
	}
	//Rows affected
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to retrieve rows affected value -> "+err.Error())
	}
	//Not found
	if rows == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("User with ID='%d' is not found", req.User.Id))
	}
	return &UpdateResponse{Api: req.Api, Updated: rows}, nil
}

func (s *userServiceServer) Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	log.Printf("Delete:\tAPI: %s\tId: %d", req.Api, req.Id)
	//SQL connection pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	//Delete user with id
	res, err := c.ExecContext(ctx, "DELETE FROM user WHERE `id`=?", req.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to delete user -> "+err.Error())
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to retrieve rows affected value -> "+err.Error())
	}

	if rows == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("user with ID='%d' is not found", req.Id))
	}
	return &DeleteResponse{Api: req.Api, Deleted: rows}, nil
}

func (s *userServiceServer) ReadAll(ctx context.Context, req *ReadAllRequest) (*ReadAllResponse, error) {
	log.Printf("Read all:\tAPI: %s", req.Api)
	//SQL connection pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	// get user list
	rows, err := c.QueryContext(ctx, "SELECT * FROM ToDo")
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to select from ToDo-> "+err.Error())
	}
	defer rows.Close()
	users := []*User{}
	for rows.Next() {
		user := new(User)
		if err := rows.Scan(&user.Id, &user.Name, &user.Phone); err != nil {
			return nil, status.Error(codes.Unknown, "Failed to retrieve field values from user row -> "+err.Error())
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, status.Error(codes.Unknown, "Failed to retrieve data from user-> "+err.Error())
	}
	return &ReadAllResponse{Api: req.Api, Users: users}, nil
}
