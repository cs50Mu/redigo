package main

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

// ConnPool is a pool of connections to redis server
type ConnPool struct {
	host string
	port string
	//maxIdle int
	maxOpen int
	// 使用中的连接数
	inUseCnt int
	// 空闲的连接
	idleList *stack
	mu       sync.Mutex
}

type stack struct {
	storage []net.Conn
	top     int
}

func newStack() *stack {
	return &stack{
		storage: make([]net.Conn, 10),
	}
}

func (s *stack) resize(n int) {
	temp := make([]net.Conn, n)
	copy(temp, s.storage[0:s.top])
	s.storage = temp
}

func (s *stack) push(e net.Conn) {
	if s.top == len(s.storage) {
		s.resize(2 * len(s.storage))
	}
	s.storage[s.top] = e
	s.top++
}

func (s *stack) pop() net.Conn {
	s.top--
	res := s.storage[s.top]
	if s.top > 0 && s.top == len(s.storage)/4 {
		s.resize(len(s.storage) / 2)
	}
	return res
}

func (s *stack) length() int {
	return s.top
}

// NewConnPool creates a new ConnPool
func NewConnPool(host, port string, maxOpen int) *ConnPool {
	return &ConnPool{
		host:     host,
		port:     port,
		idleList: newStack(),
		// sets the maximum number of connections in the idle
		// connection pool.
		//maxIdle: maxIdle,
		// sets the maximum number of open connections to the database
		maxOpen: maxOpen,
	}
}

// GetConn returns a redis connection, returns error when no free conn is available
func (cp *ConnPool) GetConn() (net.Conn, error) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	// has idle connection
	if cp.idleList.length() > 0 {
		return cp.idleList.pop(), nil
	}
	// has reached max open connnection
	if cp.inUseCnt >= cp.maxOpen {
		return nil, errors.New("exhausted pool")
	}
	// no idle connnection but hasn't reach max open connection
	cp.inUseCnt++
	return cp.connect()
}

func (cp *ConnPool) connect() (net.Conn, error) {
	address := fmt.Sprintf("%s:%s", cp.host, cp.port)
	return net.Dial("tcp", address)
}

// ReleaseConn put a connection back into pool
func (cp *ConnPool) ReleaseConn(c net.Conn) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.idleList.push(c)
	cp.inUseCnt--
	return
}
