package module

import (
	"github.com/distribyted/distribyted/config"
	"github.com/go-redis/redis"
)

var RedisClient *redis.Client

func init() {
	/*RedisClient = redis.NewClient(&redis.Options{
		Addr:     "10.88.50.113:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
		PoolSize: 20,
	})
	pong, err := RedisClient.Ping().Result()
	fmt.Println(pong, err)
	userList := make(map[string]interface{})
	userList["testuser1"] = "123456"
	userList["testuser2"] = "123456"
	RedisClient.HMSet("userList", userList)
	fmt.Println(RedisClient.HGet("userList", "testuser1").Val())*/
}

func GetUserList() []config.UserInfo {
	userList := RedisClient.HGetAll("userList").Val()
	var data []config.UserInfo
	for k, v := range userList {
		data = append(data, config.UserInfo{
			User: k,
			Pass: v,
		})
	}
	return data
}
