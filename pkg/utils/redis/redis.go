package redis

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/openinsight-proj/elastic-alert/pkg/conf"
)

const (
	AlertQueueListKey = "prom_elastic_alert:alerts:list"
)

var Client *redis.Client

func Setup() {
	conn := conf.AppConf.Redis
	redisDb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", conn.Addr, conn.Port),
		Password:     conn.Password,
		DB:           conn.Db,
		ReadTimeout:  time.Second * time.Duration(conn.ReadTimeout),
		WriteTimeout: time.Second * time.Duration(conn.WriteTimeout),
		PoolSize:     conn.PoolSize,
		DialTimeout:  time.Second * time.Duration(conn.DialTimeout),
	})
	Client = redisDb
}
