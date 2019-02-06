package redis

import (
	"fmt"

	"github.com/go-redis/redis"
)

type RedisClient struct {
	RClient *redis.Client
}

func (R *RedisClient) NewClient(address string) {
	R.RClient = redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := R.RClient.Ping().Result()
	fmt.Println(pong, err)
	// Output: PONG <nil>
}

func (R *RedisClient) SetToCache(key string, val string) error {
	err := R.RClient.Set(key, val, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (R *RedisClient) GetFromCache(key string) (string, error) {
	val, err := R.RClient.Get(key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}
