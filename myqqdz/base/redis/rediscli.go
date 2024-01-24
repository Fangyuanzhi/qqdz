package redis

import (
	"github.com/gomodule/redigo/redis"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func GetToken(length int) string {
	rand.Seed(time.Now().UnixNano())
	rs := make([]string, length)
	for start := 0; start < length; start++ {
		t := rand.Intn(3)
		if t == 0 {
			rs = append(rs, strconv.Itoa(rand.Intn(10)))
		} else if t == 1 {
			rs = append(rs, string(rand.Intn(26)+65))
		} else {
			rs = append(rs, string(rand.Intn(26)+97))
		}
	}
	return strings.Join(rs, "")
}

// Setup RedisConn redis连接池
func Setup(addr string) *redis.Pool {
	return &redis.Pool{
		//最大空闲连接数
		MaxIdle: 30,
		//在给定时间内，允许分配的最大连接数（当为零时，没有限制）
		MaxActive: 30,
		//在给定时间内，保持空闲状态的时间，若到达时间限制则关闭连接（当为零时，没有限制）
		IdleTimeout: 200,
		//提供创建和配置应用程序连接的一个函数
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			if err != nil {
				return nil, err
			}
			//如果redis设置了用户密码，使用auth指令
			//if _, err := c.Do("AUTH", password); err != nil {
			//	c.Close()
			//	return nil, err
			//}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
