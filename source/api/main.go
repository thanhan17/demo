package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/thanhan17/demo/source/api/auth"
	handlers "github.com/thanhan17/demo/source/api/handler"
	"github.com/thanhan17/demo/source/api/middleware"
	v1 "github.com/thanhan17/demo/source/grpc/model/v1"
	logging "github.com/thanhan17/demo/source/pkg/logging"
	redislib "github.com/thanhan17/demo/source/pkg/redislib"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

const (
	timeout = 100 * time.Millisecond
)

var (
	userClient v1.UserServiceClient
	log        *zap.Logger
)

func init() {
	rl, err := rotatelogs.New(
		"logs/api/%Y-%m-%d-%H",
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
	log.Info("Initing api services")
	if err := godotenv.Load(); err != nil {
		log.Error("No .env file found", zap.Error(err))
		panic("No .env file found")
	}

}

func routeUser(router *gin.RouterGroup) {
	router.GET("/", readAllUsers)
	router.GET("/:id", readUserById)
	router.POST("/", createUser)
	router.PUT("/", updateUser)
	router.DELETE("/:id", deleteUser)
}

func ginListen(srv *http.Server) {
	log.Info("Starting listen on server", zap.String("Address", srv.Addr), zap.String("service", "auth, rest api"))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("Error listen: ", zap.Error(err))
		panic("Error listen: " + err.Error())
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
	//Env
	appAddr := ":" + os.Getenv("REST_PORT")
	redis_host := os.Getenv("REDIS_HOST")
	redis_port := os.Getenv("REDIS_PORT")
	redis_password := os.Getenv("REDIS_PASSWORD")
	grpc_port := os.Getenv("GRPC_PORT")
	promAddr := os.Getenv("PROM_API_ADDR")
	//Redis
	redisClient := redislib.NewRedisDB(context.Background(), redis_host, redis_port, redis_password)
	defer redisClient.Close()

	//GRPC service
	conn, err := grpc.Dial(":"+grpc_port, grpc.WithInsecure())
	if err != nil {
		log.Error("Can't connect server gRPC", zap.Error(err))
		panic("Can't connect server gRPC" + err.Error())
	}
	defer conn.Close()
	userClient = v1.NewUserServiceClient(conn)

	//Auth Service
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
	srvAuthAndAPI := &http.Server{
		Addr:    appAddr,
		Handler: router,
	}
	go ginListen(srvAuthAndAPI)
	srvProm := &http.Server{
		Addr:    promAddr,
		Handler: promhttp.Handler(),
	}
	go promListen(srvProm)
	//Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Info("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srvAuthAndAPI.Shutdown(ctx); err != nil {
		log.Info("Server API Shutdown:", zap.Error(err))
	}
	if err := srvProm.Shutdown(ctx); err != nil {
		log.Info("Server Prometheus Shutdown:", zap.Error(err))
	}
	defer log.Info("Server exiting")
}
