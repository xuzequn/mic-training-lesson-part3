package distributed_lock

import (
	"fmt"
	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"mic-training-lesson-part3/internal"
)

func RedisLock() {

	redisAddr := fmt.Sprintf("%s:%d", internal.AppConf.RedisConfig.Host,
		internal.AppConf.RedisConfig.Port)
	// 客户端
	client := goredislib.NewClient(&goredislib.Options{
		Addr: redisAddr,
	})
	// 连接池
	pool := goredis.NewPool(client)
	// redis 分布式锁
	rs := redsync.New(pool)

}
