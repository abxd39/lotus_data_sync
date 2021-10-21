package utils

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

func RedisInit() {
	/**
	文档地址：
	https://godoc.org/github.com/go-redis/redis
	*/
	addr := Initconf.String("RedisHost")
	Rdb16 = redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
		DB:           15,
	})
	_, err := Rdb16.Ping().Result()
	if err != nil {
		Log.Errorf("Run connect  readisio is failed：%v ", err)
		panic(fmt.Sprintf("连接redis失败，终止启动，err=%v", err))
	} else {
		Log.Traceln("redis init success")
	}

}
