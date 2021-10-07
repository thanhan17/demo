package v1

import (
	"context"
	"database/sql"
	fmt "fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	apiVersion = "v1"
	mutexUser  = "userId"
	layoutDate = "20060102"
)

type userServiceServer struct {
	db          *sql.DB
	redisClient *redis.Client
}

func NewUserServiceServer(db *sql.DB, redisClient *redis.Client) UserServiceServer {
	return &userServiceServer{db: db, redisClient: redisClient}
}

func (s *userServiceServer) connectDB(ctx context.Context) (*sql.Conn, error) {
	c, err := s.db.Conn(ctx)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to connect to database-> "+err.Error())
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

func getUserId(ctx context.Context, redisClient *redis.Client, t string) (string, error) {
	key := "user" + t
	//Redsync
	pool := goredis.NewPool(redisClient)
	rs := redsync.New(pool)
	mutex := rs.NewMutex(mutexUser)
	if err := mutex.Lock(); err != nil {
		return "", status.Error(codes.Unknown, "Failed to create lock mutex userId -> "+err.Error())
	}
	//Increase key
	_, err := redisClient.Incr(ctx, key).Result()
	if err != nil {
		return "", err
	}
	//Get key
	userId, err := redisClient.Get(ctx, "user"+t).Result()
	if err != nil {
		return "", err
	}
	if ok, err := mutex.Unlock(); !ok || err != nil {
		return "", status.Error(codes.Unknown, "Failed to unlock mutex userId -> "+err.Error())
	}
	return userId, nil
}

func lpad(s string, pad string, plength int) string {
	for i := len(s); i < plength; i++ {
		s = pad + s
	}
	return s
}

// Not check exist and validation after insert
func (s *userServiceServer) Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error) {
	log.Printf("Create: " + req.String())
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	//SQL connection pool
	c, err := s.connectDB(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	t := time.Now().Format(layoutDate)
	//Redis mutex
	userId, err := getUserId(ctx, s.redisClient, t)
	if err != nil {
		return nil, err
	}
	userId = t + lpad(userId, "0", 8)
	//Insert entity user data
	_, err = c.ExecContext(ctx, "INSERT INTO user(`id`, `name`, `phone`) values (?, ?, ?)", userId, req.User.Name, req.User.Phone)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to insert into user -> "+err.Error())
	}
	return &CreateResponse{Api: req.Api, Id: userId}, nil
}

func (s *userServiceServer) Read(ctx context.Context, req *ReadRequest) (*ReadResponse, error) {
	log.Printf("Read: " + req.String())
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}
	//SQL connection pool
	c, err := s.connectDB(ctx)
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
	var created_at, updated_at time.Time
	if row.Next() {
		if err := row.Scan(&user.Id, &user.Name, &user.Phone, &created_at, &updated_at); err != nil {
			return nil, status.Error(codes.NotFound, "Failed to retrieve field values from user -> "+err.Error())
		}
	}
	//Format timestamp
	user.CreatedAt = timestamppb.New(created_at)
	user.UpdateAt = timestamppb.New(updated_at)
	if row.Next() {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("Found multiple user rows with id='%s'", req.Id))
	}
	return &ReadResponse{Api: req.Api, User: &user}, nil
}

func (s *userServiceServer) Update(ctx context.Context, req *UpdateRequest) (*UpdateResponse, error) {
	log.Println("Update: ", req.String())
	//SQL connection pool
	c, err := s.connectDB(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	// update ToDo
	res, err := c.ExecContext(ctx, "UPDATE user SET `name`=?, `phone`=? WHERE `id`=?",
		req.User.Name, req.User.Phone, req.User.Id)
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
		return nil, status.Error(codes.NotFound, fmt.Sprintf("User with ID='%s' is not found", req.User.Id))
	}
	return &UpdateResponse{Api: req.Api, Updated: rows}, nil
}

func (s *userServiceServer) Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	log.Println("Delete: ", req.String())
	//SQL connection pool
	c, err := s.connectDB(ctx)
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
		return nil, status.Error(codes.NotFound, fmt.Sprintf("user with ID='%s' is not found", req.Id))
	}
	return &DeleteResponse{Api: req.Api, Deleted: rows}, nil
}

func (s *userServiceServer) ReadAll(ctx context.Context, req *ReadAllRequest) (*ReadAllResponse, error) {
	log.Println("Read all: ", req.String())
	//SQL connection pool
	c, err := s.connectDB(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	// get user list
	rows, err := c.QueryContext(ctx, "SELECT * FROM user")
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to select from ToDo-> "+err.Error())
	}
	defer rows.Close()
	users := []*User{}
	var created_at, updated_at time.Time
	for rows.Next() {
		user := new(User)
		if err := rows.Scan(&user.Id, &user.Name, &user.Phone, &created_at, &updated_at); err != nil {
			return nil, status.Error(codes.Unknown, "Failed to retrieve field values from user row -> "+err.Error())
		}
		user.CreatedAt = timestamppb.New(created_at)
		user.UpdateAt = timestamppb.New(updated_at)
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, status.Error(codes.Unknown, "Failed to retrieve data from user-> "+err.Error())
	}
	return &ReadAllResponse{Api: req.Api, Users: users}, nil
}
