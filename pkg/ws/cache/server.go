// Package cache 缓存
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/redis/go-redis/v9"

	"bingo/pkg/ws/model"
)

var (
	RedisClient *redis.Client
)

const (
	serversHashKey       = "acc:hash:servers" // 全部的服务器
	serversHashCacheTime = 2 * 60 * 60        // key过期时间
	serversHashTimeout   = 3 * 60             // 超时时间
)

func getServersHashKey() (key string) {
	key = fmt.Sprintf("%s", serversHashKey)

	return
}

// SetServerInfo 设置服务器信息
func SetServerInfo(server *model.Server, currentTime uint64) (err error) {
	key := getServersHashKey()
	value := fmt.Sprintf("%d", currentTime)
	number, err := RedisClient.Do(context.Background(), "hSet", key, server.String(), value).Int()
	if err != nil {
		log.Println("SetServerInfo", key, number, err)

		return
	}

	RedisClient.Do(context.Background(), "Expire", key, serversHashCacheTime)

	return
}

// DelServerInfo 下线服务器信息
func DelServerInfo(server *model.Server) (err error) {
	key := getServersHashKey()
	number, err := RedisClient.Do(context.Background(), "hDel", key, server.String()).Int()
	if err != nil {
		log.Println("DelServerInfo", key, number, err)

		return
	}

	if number != 1 {
		return
	}

	RedisClient.Do(context.Background(), "Expire", key, serversHashCacheTime)

	return
}

// GetServerAll 获取所有服务器
func GetServerAll(currentTime uint64) (servers []*model.Server, err error) {
	servers = make([]*model.Server, 0)
	key := getServersHashKey()

	val, err := RedisClient.Do(context.Background(), "hGetAll", key).Result()
	valByte, _ := json.Marshal(val)

	log.Println("GetServerAll", key, string(valByte))

	serverMap, err := RedisClient.HGetAll(context.Background(), key).Result()
	if err != nil {
		log.Println("SetServerInfo", key, err)

		return
	}

	for key, value := range serverMap {
		valueUint64, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			log.Println("GetServerAll", key, err)

			return nil, err
		}

		// 超时
		if valueUint64+serversHashTimeout <= currentTime {
			continue
		}

		server, err := model.StringToServer(key)
		if err != nil {
			log.Println("GetServerAll", key, err)

			return nil, err
		}

		servers = append(servers, server)
	}

	return
}
