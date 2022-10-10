package main

import "myredis/redislock"

func main(){
	go redislock.LockMq("helloworld")
	select{}
}
