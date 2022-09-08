package distributed_lock

import (
	"sync"
	"testing"
)

func TestRedisLock(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go RedisLock(&wg)
	}
	wg.Wait()
}
