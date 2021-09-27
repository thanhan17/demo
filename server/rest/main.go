package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/thanhan17/demo/auth"
	v1 "github.com/thanhan17/demo/grpc/model/v1"
	handlers "github.com/thanhan17/demo/handler"
	"github.com/thanhan17/demo/middleware"
	"google.golang.org/grpc"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

const (
	timeout = 100 * time.Millisecond
)

var (
	userClient v1.UserServiceClient
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
		panic("No .env file found")
	}
}

func newRedisDB(ctx context.Context, host, port, password string) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       0,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		panic("Unable to connect to redis " + err.Error())
	}
	return redisClient
}

func routeUser(router *gin.RouterGroup) {
	router.GET("/", readAllUsers)
	router.GET("/:id", readUserById)
	router.POST("/", createUser)
	router.PUT("/", updateUser)
	router.DELETE("/:id", deleteUser)
}

func main() {
	//Env
	appAddr := ":" + os.Getenv("REST_PORT")
	redis_host := os.Getenv("REDIS_HOST")
	redis_port := os.Getenv("REDIS_PORT")
	redis_password := os.Getenv("REDIS_PASSWORD")
	grpc_port := os.Getenv("GRPC_PORT")

	//Redis
	redisClient := newRedisDB(context.Background(), redis_host, redis_port, redis_password)
	defer redisClient.Close()

	//GRPC service
	conn, err := grpc.Dial(":"+grpc_port, grpc.WithInsecure())
	if err != nil {
		panic("Can't connect server gRPC" + err.Error())
	}
	defer conn.Close()
	userClient = v1.NewUserServiceClient(conn)

	//Service
	var rd = auth.NewAuth(redisClient)
	var tk = auth.NewToken()
	var service = handlers.NewProfile(rd, tk)

	//Handle
	var router = gin.Default()
	router.POST("/login", service.Login)
	router.POST("/logout", middleware.TokenAuthMiddleware(), service.Logout)
	router.POST("/refresh", service.Refresh)
	api := router.Group("/api/:api/user")
	routeUser(api)

	//Create http server, handle and listen at port
	srv := &http.Server{
		Addr:    appAddr,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error listen: %s\n", err)
		}
	}()

	//Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
