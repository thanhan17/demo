package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	v1 "github.com/thanhan17/demo/grpc/model/v1"
	"google.golang.org/grpc/status"
)

func response(c *gin.Context, data interface{}, err error) {
	statusCode := http.StatusOK
	var errorMessage string
	if err != nil {
		log.Println("Server Error Occured:", err)
		errorMessage = strings.Title(err.Error())
		statusCode = http.StatusInternalServerError
	}
	c.JSON(statusCode, gin.H{"data": data, "error": errorMessage})
}

func getParam(c *gin.Context, param string) (string, error) {
	p := c.Param(param)
	if len(p) == 0 {
		return "", errors.New("invalid parameter: " + p)
	}
	return p, nil
}

func readAllUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, timeout*time.Second)
	defer cancel()
	data, err := userClient.ReadAll(ctx, &v1.ReadAllRequest{Api: "1"})
	if err != nil {
		response(c, nil, errors.New(status.Convert(err).Message()))
	}
	response(c, data, err)
}

func readUserById(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, timeout*time.Second)
	defer cancel()

	strId, err := getParam(c, "id")
	if err != nil {
		response(c, nil, err)
		return
	}

	id, err := strconv.ParseInt(strId, 10, 64)
	if err != nil {
		response(c, nil, err)
		return
	}

	data, err := userClient.Read(ctx, &v1.ReadRequest{Api: "1", Id: id})
	if err != nil {
		response(c, nil, errors.New(status.Convert(err).Message()))
	}
	response(c, data, err)
}

func CreateUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, timeout*time.Second)
	defer cancel()

	var user v1.User
	err := c.BindJSON(&user)
	if err != nil {
		response(c, nil, err)
		return
	}

	data, err := userClient.Create(ctx, &v1.CreateRequest{Api: "1", User: &user})
	response(c, data, err)
}

func UpdateUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, timeout)
	defer cancel()

	var user v1.User
	err := c.BindJSON(&user)
	if err != nil {
		response(c, nil, err)
		return
	}

	data, err := userClient.Update(ctx, &v1.UpdateRequest{Api: "1", User: &user})
	response(c, data, err)
}

func DeleteUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, timeout*time.Second)
	defer cancel()

	strId, err := getParam(c, "id")
	if err != nil {
		response(c, nil, err)
		return
	}

	id, err := strconv.ParseInt(strId, 10, 64)
	if err != nil {
		response(c, nil, err)
		return
	}

	data, err := userClient.Delete(ctx, &v1.DeleteRequest{Api: "1", Id: id})
	if err != nil {
		response(c, nil, errors.New(status.Convert(err).Message()))
	}
	response(c, data, err)
}
