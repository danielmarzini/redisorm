package redisorm

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/garyburd/redigo/redis"
)

type RedisConf struct {
	maxIdle   int
	maxActive int
	port      string
	ip        string
}
type Redis struct {
	initialized bool
	conf        *RedisConf
}

var std = &Redis{
	initialized: true,
	conf: &RedisConf{
		maxIdle:   0,
		maxActive: 0,
		port:      "6379",
		ip:        "127.0.0.1",
	},
}

var pool *redis.Pool

func (r *Redis) Configure(cnf map[string]interface{}) error {
	ok := false
	if r.conf.maxIdle, ok = cnf["poolMaxIdle"].(int); !ok {
		return errors.New("redis pool max idle not specified")
	}
	if r.conf.maxActive, ok = cnf["poolMaxActive"].(int); !ok {
		return errors.New("redis pool max active not specified")
	}
	if r.conf.port, ok = cnf["port"].(string); !ok {
		return errors.New("redis port not specified")
	}
	if r.conf.ip, ok = cnf["ip"].(string); !ok {
		return errors.New("redis ip not specified")
	}
	r.initialized = true
	pool = r.newPool()
	return nil
}

func Configure(cnf map[string]interface{}) error {
	if err := std.Configure(cnf); err != nil {
		return err
	}
	return nil
}

func (r *Redis) newPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:   r.conf.maxIdle,
		MaxActive: r.conf.maxActive, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", r.conf.ip+":"+r.conf.port)
			if err != nil {
				return c, err
			}
			return c, err
		},
	}
}

func Get(key string, obj interface{}) error {
	c := pool.Get()
	defer c.Close()
	jsonString, err := redis.String(c.Do("GET", key))
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(jsonString), obj)
	if err != nil {
		return err
	}
	return nil
}

func List(key string, objs []interface{}) error {
	c := pool.Get()
	defer c.Close()

	jsonString, err := redis.String(c.Do("KEYS", key))
	if err != nil {
		return err
	}
	fmt.Println(jsonString)

	return nil
}

func Set(key string, obj interface{}) error {
	c := pool.Get()
	defer c.Close()

	jsonString, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	_, err = redis.String(c.Do("SET", key, string(jsonString[:len(jsonString)])))
	if err != nil {
		return err
	}
	return nil
}
