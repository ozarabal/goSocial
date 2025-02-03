package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/ozarabal/goSocial/internal/store"
)

type UserStore struct {
	rdb *redis.Client
}

const UserExpTime = time.Minute

func (s *UserStore) Get(ctx context.Context, userID int64) (*store.User,error){
	cacheKey := fmt.Sprintf("user-%d", userID)

	data, err:= s.rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil{
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var user store.User
	if data != "" {
		fmt.Println("masuk sini")
		err := json.Unmarshal([]byte(data), &user)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil 
}

func (s *UserStore) Set(ctx context.Context, user *store.User) error {
	cachekey := fmt.Sprintf("user-%d", user.ID)

	json, err := json.Marshal(user)
	if err != nil{
		return err
	}

	return s.rdb.SetEX(ctx, cachekey, json, UserExpTime).Err()
}

func (s *UserStore) Delete(ctx context.Context, userID int64){
	cachekey := fmt.Sprintf("user-%d",userID)
	s.rdb.Del(ctx,cachekey)
}