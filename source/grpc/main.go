package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	v1 "github.com/thanhan17/demo/source/grpc/model/v1"
	logging "github.com/thanhan17/demo/source/pkg/logging"
	redislib "github.com/thanhan17/demo/source/pkg/redislib"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	layoutDate = "20060102"
)

var (
	log *zap.Logger
)

func init() {
	rl, err := rotatelogs.New(
		"logs/grpc/%Y-%m-%d-%H",
		rotatelogs.WithClock(rotatelogs.Local),
		rotatelogs.WithRotationTime(time.Hour),
	)
	if err != nil {
		panic("Can't create logger")
	}
	core := zapcore.NewTee(
		zapcore.NewCore(logging.GetJSONEncoderZap(), zapcore.AddSync(rl), zap.DebugLevel),
		zapcore.NewCore(logging.GetConsoleEncoderZap(), zapcore.AddSync(os.Stdout), zap.InfoLevel),
	)
	log = zap.New(core)
	log.Info("Initing grpc services")
	if err := godotenv.Load(); err != nil {
		log.Error("No .env file found", zap.Error(err))
		panic("No .env file found")
	}
}

func promListen(srv *http.Server) {
	log.Info("Starting listen on server", zap.String("Address", srv.Addr), zap.String("service", "monitoring"))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("Error listen: ", zap.Error(err))
		panic("Error listen: " + err.Error())
	}
}

func main() {
	ctx := context.Background()
	port := os.Getenv("GRPC_PORT")
	dbUser := os.Getenv("DBUser")
	dbPassword := os.Getenv("DBPassword")
	dbHost := os.Getenv("DBHost")
	dbName := os.Getenv("DBName")
	promAddr := os.Getenv("PROM_GRPC_ADDR")
	param := "parseTime=true"
	config := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", dbUser, dbPassword, dbHost, dbName, param)
	db, err := sql.Open("mysql", config)
	if err != nil {
		log.Error("failed to open database: %v", zap.Error(err))
		return
	}
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(50)
	db.SetConnMaxLifetime(time.Second * 5)
	defer db.Close()

	//Prometheus
	srvProm := &http.Server{
		Addr:    promAddr,
		Handler: promhttp.Handler(),
	}
	go promListen(srvProm)
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
		log.Info("Created redis key user" + t)
	} else if err != nil {
		panic(err)
	}
	v1API := v1.NewUserServiceServer(db, redisClient)
	if err := runServer(ctx, v1API, port); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
