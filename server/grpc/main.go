package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/thanhan17/demo/grpc"
	v1 "github.com/thanhan17/demo/grpc/model/v1"
	redislib "github.com/thanhan17/demo/lib/redislib"
)

const (
	layoutDate = "20060102"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
		panic("No .env file found")
	}
}

func main() {
	ctx := context.Background()
	port := os.Getenv("GRPC_PORT")
	dbUser := os.Getenv("DBUser")
	dbPassword := os.Getenv("DBPassword")
	dbHost := os.Getenv("DBHost")
	dbName := os.Getenv("DBName")
	param := "parseTime=true"
	config := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", dbUser, dbPassword, dbHost, dbName, param)
	db, err := sql.Open("mysql", config)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
		return
	}
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(50)
	db.SetConnMaxLifetime(time.Second * 5)
	defer db.Close()

	//Env
	redis_host := os.Getenv("REDIS_HOST")
	redis_port := os.Getenv("REDIS_PORT")
	redis_password := os.Getenv("REDIS_PASSWORD")

	//Redis
	redisClient := redislib.NewRedisDB(context.Background(), redis_host, redis_port, redis_password)
	defer redisClient.Close()
	t := time.Now().Format(layoutDate)
	_, err = redisClient.Get(ctx, "user"+t).Result()
	if err == redis.Nil {
		yyyy, mm, dd := time.Now().AddDate(0, 0, 1).Date()
		_, err = redisClient.Set(ctx, "user"+t, "0", time.Until(time.Date(yyyy, mm, dd, 0, 0, 0, 0, time.Now().Location()))).Result()
		if err != nil {
			panic(err)
		}
		log.Println("Created redis key user" + t)
	} else if err != nil {
		panic(err)
	}
	v1API := v1.NewUserServiceServer(db, redisClient)
	if err := grpc.RunServer(ctx, v1API, port); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
