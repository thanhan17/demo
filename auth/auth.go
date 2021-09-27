package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type AuthInterface interface {
	CreateAuth(context.Context, string, *TokenDetails) error
	FetchAuth(context.Context, string) (string, error)
	DeleteRefresh(context.Context, string) error
	DeleteTokens(context.Context, *AccessDetails) error
}

type service struct {
	client *redis.Client
}

var _ AuthInterface = &service{}

func NewAuth(client *redis.Client) *service {
	return &service{client: client}
}

type AccessDetails struct {
	TokenUuid string
	UserId    string
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	TokenUuid    string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

//Save token metadata to Redis
func (tk *service) CreateAuth(ctx context.Context, userId string, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	atCreated, err := tk.client.Set(ctx, td.TokenUuid, userId, at.Sub(now)).Result()
	if err != nil {
		return err
	}
	rtCreated, err := tk.client.Set(ctx, td.RefreshUuid, userId, rt.Sub(now)).Result()
	if err != nil {
		return err
	}
	if atCreated == "0" || rtCreated == "0" {
		return errors.New("no record inserted")
	}
	return nil
}

//Check the metadata saved
func (tk *service) FetchAuth(ctx context.Context, tokenUuid string) (string, error) {
	userid, err := tk.client.Get(ctx, tokenUuid).Result()
	if err != nil {
		return "", err
	}
	return userid, nil
}

//Once a user row in the token table
func (tk *service) DeleteTokens(ctx context.Context, authD *AccessDetails) error {
	//get the refresh uuid
	refreshUuid := fmt.Sprintf("%s++%s", authD.TokenUuid, authD.UserId)
	//delete access token
	deletedAt, err := tk.client.Del(ctx, authD.TokenUuid).Result()
	if err != nil {
		return err
	}
	//delete refresh token
	deletedRt, err := tk.client.Del(ctx, refreshUuid).Result()
	if err != nil {
		return err
	}
	//When the record is deleted, the return value is 1
	if deletedAt != 1 || deletedRt != 1 {
		return errors.New("something went wrong")
	}
	return nil
}

func (tk *service) DeleteRefresh(ctx context.Context, refreshUuid string) error {
	//delete refresh token
	deleted, err := tk.client.Del(ctx, refreshUuid).Result()
	if err != nil || deleted == 0 {
		return err
	}
	return nil
}
