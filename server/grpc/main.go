package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/thanhan17/demo/grpc"
	v1 "github.com/thanhan17/demo/grpc/model/v1"
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
	db.SetMaxOpenConns(30)
	db.SetMaxIdleConns(30)
	db.SetConnMaxLifetime(time.Second * 5)
	defer db.Close()
	v1API := v1.NewUserServiceServer(db)
	if err := grpc.RunServer(ctx, v1API, port); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
