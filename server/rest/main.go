package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/thanhan17/demo/auth"
	handlers "github.com/thanhan17/demo/handler"
	"github.com/thanhan17/demo/middleware"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
		panic("No .env file found")
	}
}

func NewRedisDB(host, port, password string) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       0,
	})
	return redisClient
}

func main() {
	//Env
	appAddr := ":" + os.Getenv("REST_PORT")
	redis_host := os.Getenv("REDIS_HOST")
	redis_port := os.Getenv("REDIS_PORT")
	redis_password := os.Getenv("REDIS_PASSWORD")

	//Redis
	redisClient := NewRedisDB(redis_host, redis_port, redis_password)
	defer redisClient.Close()

	//Service
	var rd = auth.NewAuth(redisClient)
	var tk = auth.NewToken()
	var service = handlers.NewProfile(rd, tk)

	//Handle
	var router = gin.Default()
	router.POST("/login", service.Login)
	router.POST("/todo", middleware.TokenAuthMiddleware(), service.CreateTodo)
	router.POST("/logout", middleware.TokenAuthMiddleware(), service.Logout)
	router.POST("/refresh", service.Refresh)

	//Create http server, handle and listen at port
	srv := &http.Server{
		Addr:    appAddr,
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
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
