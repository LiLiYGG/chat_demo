package cache

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
	logging "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

// RedisClient Redis缓存客户端单例
var (
	RedisClient *redis.Client
	RedisDb     string
	RedisAddr   string
	RedisPw     string
	RedisDbName string
)

// Redis 初始化redis链接
func init() {
	// 加载配置信息
	file, err := ini.Load("E:\\chat-demo\\conf\\conf.ini")
	if err != nil {
		fmt.Println("Redis 配置文件读取错误，请检查文件路径:", err)
	}
	// 读取配置信息
	LoadRedisData(file)
	Redis()
}

// Redis 在中间件中初始化redis链接
func Redis() {
	db, _ := strconv.ParseUint(RedisDbName, 10, 64)
	client := redis.NewClient(&redis.Options{
		Addr: RedisAddr,
		DB:   int(db),
	})
	_, err := client.Ping().Result()
	if err != nil {
		logging.Info(err)
		panic(err)
	}
	RedisClient = client
}

func LoadRedisData(file *ini.File) {
	RedisDb = file.Section("redis").Key("RedisDb").String()
	RedisAddr = file.Section("redis").Key("RedisAddr").String()
	RedisPw = file.Section("redis").Key("RedisPw").String()
	RedisDbName = file.Section("redis").Key("RedisDbName").String()
}
