package distributed_lock

import (
	"fmt"
	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"sync"
	"time"
)

func RedisLock(wg *sync.WaitGroup) {

	//redisAddr := fmt.Sprintf("%s:%d", internal.AppConf.RedisConfig.Host,
	//internal.AppConf.RedisConfig.Port)
	redisAddr := fmt.Sprintf("%s:%d", "127.0.0.1", 6379)
	// 客户端
	client := goredislib.NewClient(&goredislib.Options{
		Addr:     redisAddr,
		Password: "#foVEYdich",
	})
	// 连接池
	pool := goredis.NewPool(client)
	// redis 分布式锁
	rs := redsync.New(pool)

	mutexname := "product@1"
	mutex := rs.NewMutex(mutexname, redsync.WithExpiry(30*time.Second))
	fmt.Println("Lock()....")
	err := mutex.Lock()
	if err != nil {
		panic(err)
	}
	// 业务逻辑
	fmt.Println("Get Lock!!!")
	time.Sleep(time.Second * 1)

	fmt.Println("UnLock()")
	ok, err := mutex.Unlock()
	if !ok || err != nil {
		panic(err)
	}
	fmt.Println("Released Lock!!!")

}
