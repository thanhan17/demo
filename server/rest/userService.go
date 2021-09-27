package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

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
	ctx, cancel := context.WithTimeout(c, timeout)
	defer cancel()

	api, err := getParam(c, "api")
	if err != nil {
		response(c, nil, err)
		return
	}

	data, err := userClient.ReadAll(ctx, &v1.ReadAllRequest{Api: api})
	if err != nil {
		response(c, nil, errors.New(status.Convert(err).Message()))
	}
	response(c, data, err)
}

func readUserById(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, timeout)
	defer cancel()

	api, err := getParam(c, "api")
	if err != nil {
		response(c, nil, err)
		return
	}

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

	data, err := userClient.Read(ctx, &v1.ReadRequest{Api: api, Id: id})
	if err != nil {
		response(c, nil, errors.New(status.Convert(err).Message()))
	}
	response(c, data, err)
}

func createUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, timeout)
	defer cancel()
	api, err := getParam(c, "api")
	if err != nil {
		response(c, nil, err)
		return
	}

	var user v1.User
	if err = c.ShouldBindJSON(&user); err != nil {
		response(c, nil, err)
		return
	}

	data, err := userClient.Create(ctx, &v1.CreateRequest{Api: api, User: &user})
	log.Println(data.String())
	response(c, data, err)
}

func updateUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, timeout)
	defer cancel()

	api, err := getParam(c, "api")
	if err != nil {
		response(c, nil, err)
		return
	}

	var user v1.User
	if err := c.ShouldBindJSON(&user); err != nil {
		response(c, nil, err)
		return
	}

	data, err := userClient.Update(ctx, &v1.UpdateRequest{Api: api, User: &user})
	response(c, data, err)
}

func deleteUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, timeout)
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
