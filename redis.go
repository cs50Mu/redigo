package main

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
)

// RedisClient represent a redis client conn
type RedisClient struct {
	pool *ConnPool
}

// NewRedisClient returns a new Redis client
func NewRedisClient(host, port string) (*RedisClient, error) {
	pool := NewConnPool(host, port, 10*runtime.NumCPU())
	return &RedisClient{
		pool: pool,
	}, nil
}

func (rc *RedisClient) executeCommand(command string, args ...string) (*Reply, error) {
	c, err := rc.pool.GetConn()
	if err != nil {
		return nil, err
	}
	defer rc.pool.ReleaseConn(c)
	commands := append([]string{command}, args...)
	err = c.SendCommand(commands...)
	if err != nil {
		return nil, err
	}
	reply, err := c.ReadResp()
	if err != nil {
		return nil, err
	}
	return reply, nil
}

// Get the value of key. If the key does not exist the special value nil is returned.
// An error is returned if the value stored at key is not a string, because GET only handles string values.
func (rc *RedisClient) Get(key string) ([]byte, error) {
	reply, err := rc.executeCommand("GET", key)
	if err != nil {
		return nil, err
	}
	return reply.stringVal, nil
}

// Set key to hold the string value. If key already holds a value, it is overwritten, regardless of its type.
// Any previous time to live associated with the key is discarded on successful SET operation.
func (rc *RedisClient) Set(key, val string) (bool, error) {
	reply, err := rc.executeCommand("SET", key, val)
	if err != nil {
		return false, err
	}
	if string(reply.stringVal) == "OK" {
		return true, nil
	}
	fmt.Printf("reply: %s\n", reply.stringVal)
	return false, err
}

// Expire set a timeout on key
// return true if the timeout was set.
// return false if key does not exist
func (rc *RedisClient) Expire(key string, sec int) (bool, error) {
	reply, err := rc.executeCommand("EXPIRE", key, strconv.Itoa(sec))
	if err != nil {
		return false, err
	}
	if reply.integerVal == 1 {
		return true, nil
	}
	return false, nil
}

// TTL Returns the remaining time to live of a key that has a timeout.
func (rc *RedisClient) TTL(key string) (int64, error) {
	reply, err := rc.executeCommand("TTL", key)
	if err != nil {
		return 0, err
	}
	if reply.integerVal >= 0 {
		return reply.integerVal, nil
	} else if reply.integerVal == -1 {
		return 0, errors.New("key exists but has no associated expire")
	} else if reply.integerVal == -2 {
		return 0, errors.New("key does not exist")
	}
	return 0, errors.New("unknown error")
}

// Keys returns all keys matching pattern
// returns nil when nothing matches the pattern
func (rc *RedisClient) Keys(pattern string) ([]string, error) {
	reply, err := rc.executeCommand("KEYS", pattern)
	if err != nil {
		return nil, err
	}
	if reply.arrayVal != nil {
		res := make([]string, 0)
		for _, r := range reply.arrayVal {
			res = append(res, string(r.stringVal))
		}
		return res, nil
	}
	return nil, nil
}

// Select the Redis logical database having the specified zero-based numeric index.
func (rc *RedisClient) Select(index int) (bool, error) {
	reply, err := rc.executeCommand("SELECT", strconv.Itoa(index))
	if err != nil {
		return false, err
	}
	if string(reply.stringVal) == "OK" {
		return true, nil
	}
	return false, nil
}

// Mset sets the given keys to their respective values.
func (rc *RedisClient) Mset(kvs map[string]string) error {
	kvList := make([]string, 0)
	for k, v := range kvs {
		kvList = append(kvList, k, v)
	}
	_, err := rc.executeCommand("MSET", kvList...)
	if err != nil {
		return err
	}
	return nil
}

// Mget Returns the values of all specified keys.
// For every key that does not hold a string value or does not exist, the special value nil is returned.
func (rc *RedisClient) Mget(keys ...string) ([][]byte, error) {
	reply, err := rc.executeCommand("MGET", keys...)
	if err != nil {
		return nil, err
	}
	res := make([][]byte, 0)
	for _, a := range reply.arrayVal {
		res = append(res, a.stringVal)
	}
	return res, nil
}

// Incr increments the number stored at key by one
func (rc *RedisClient) Incr(key string) (int64, error) {
	reply, err := rc.executeCommand("INCR", key)
	if err != nil {
		return 0, err
	}
	return reply.integerVal, nil
}

// IncrBy increments the number stored at key by increment
func (rc *RedisClient) IncrBy(key string, inc int64) (int64, error) {
	reply, err := rc.executeCommand("INCRBY", key, strconv.FormatInt(inc, 10))
	if err != nil {
		return 0, err
	}
	return reply.integerVal, nil
}

// IncrByFloat increment the string representing a floating point number stored at key by the specified increment
func (rc *RedisClient) IncrByFloat(key string, inc float64) (float64, error) {
	reply, err := rc.executeCommand("INCRBYFLOAT", key, strconv.FormatFloat(inc, 'f', 17, 64))
	if err != nil {
		return 0, err
	}
	floatVal, _ := strconv.ParseFloat(string(reply.stringVal), 64)
	return floatVal, nil
}

// Scan incrementally iterate over a collection of keys
func (rc *RedisClient) Scan(cursor int64, pattern string, count int64) (int64, []string, error) {
	args := []string{strconv.FormatInt(cursor, 10)}
	if pattern != "" {
		args = append(args, "MATCH", pattern)
	}
	if count != 0 {
		args = append(args, "COUNT", strconv.FormatInt(count, 10))
	}
	reply, err := rc.executeCommand("SCAN", args...)
	if err != nil {
		return 0, nil, err
	}
	nextCursor, _ := strconv.ParseInt(string(reply.arrayVal[0].stringVal), 10, 64)
	keys := make([]string, 0)
	for _, k := range reply.arrayVal[1].arrayVal {
		keys = append(keys, string(k.stringVal))
	}
	return nextCursor, keys, nil
}

// Del Removes the specified keys. A key is ignored if it does not exist.
func (rc *RedisClient) Del(keys ...string) (int64, error) {
	reply, err := rc.executeCommand("DEL", keys...)
	if err != nil {
		return 0, err
	}
	return reply.integerVal, nil
}

// Pipeline returns a redis pipeline
func (rc *RedisClient) Pipeline() (*Pipeline, error) {
	c, err := rc.pool.GetConn()
	if err != nil {
		return nil, err
	}
	defer rc.pool.ReleaseConn(c)
	return &Pipeline{
		conn: c,
		pool: rc.pool,
	}, nil
}
