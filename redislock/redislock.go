package redislock

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "123456",
	DB:       0,
})

func LockMq(svrName string) {
	key := fmt.Sprintf("%s_lockmq", svrName)
	// 尝试加锁
	var set bool
	for {
		set = redisLock(key, svrName, time.Second*30)
		if set {
			log.Println("redisLock success ")
			if err := afterLockSuccess(key); err != nil {
				// 如果此处有err ，自然是 mq 初始化失败
				log.Println("mq init error: ", err)
			}else{
				log.Println("redisLock expire failed ")
			}
			time.Sleep(time.Second * 10)
			continue
		}
		// afterLockFailed()
		log.Println("redisLock failed ")
		time.Sleep(time.Second * 10)
	}
}

func afterLockSuccess(key string) error {
	// 初始化需要做的内容或者句柄
	// xxx
	// 对于此处的初始化 mq 句柄失败才返回 err
	ch := make(chan struct{}, 1)
	go func() {
		// 模拟消费消息
		for {
			select {
			case <-ch:
				log.Println("expire failed，mq close")
				return
			default:
				log.Println("is consuming msg")
				time.Sleep(time.Second * 2)
			}
		}
	}()

	for {
		time.Sleep(time.Second*10)
		// 续期
		set, err := rdb.PExpire(context.TODO(), key, time.Second*30).Result()
		if err != nil {
			log.Println("PExpire error!! ", err)
			return nil
		}
		if !set {
			ch <- struct{}{}
			log.Println("PExpire failed!!")
			return nil
		}
		log.Println("PExpire success!! ")
	}

}

func redisLock(key, value string, duration time.Duration) bool {
	set, err := rdb.SetNX(context.TODO(), key, value, duration).Result()
	if err != nil {
		log.Println("setnx failed, error: ", err)
		return false
	}
	return set
}
