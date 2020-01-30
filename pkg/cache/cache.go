package cache

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"sync"
)

type Cache interface {
	GetKey(key string) (interface{}, error)
	SetKey(key string, value interface{}, expiry string) (bool, error)
}

type cache struct {
	m sync.Mutex
	c redis.Conn
}

func Init(address string, port string) Cache {
	log.Println("REDIS:: " + address + ":" + port)
	c, err := redis.Dial("tcp", address+":"+port)
	log.Println("DIAL OK")
	if err != nil {
		log.Fatal(err.Error())
	}
	return &cache{
		c: c,
	}
}

func (ch *cache) GetKey(key string) (interface{}, error) {
	s, err := ch.c.Do("GET", key)
	return s, err
}

func (ch *cache) SetKey(key string, value interface{}, expiry string) (bool, error) {
	if _, err := ch.c.Do("HMSET", redis.Args{}.Add(key).AddFlat(value)...); err != nil {
		log.Println(err)
		return false, err
	}

	if _, err := ch.c.Do(fmt.Sprintf("EXPIRE %s %s", key, expiry)); err != nil {
		log.Println(err)
		return false, err
	}
	return true, nil
}
