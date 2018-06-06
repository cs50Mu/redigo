package main

import (
	"fmt"
	"net"
	"testing"
)

func TestPush(t *testing.T) {
	fmt.Println("testing conn_pool")
	s := newStack()
	var c net.Conn
	s.push(c)
	if s.length() != 1 {
		t.Errorf("test failed, expected: 1, got: %d", s.length())
	}
}

func TestConnPool(t *testing.T) {
	connPool := NewConnPool("127.0.0.1", "6379", 10)
	for i := 1; i <= 10; i++ {
		conn, err := connPool.GetConn()
		if err != nil {
			fmt.Printf("got error: %s\n", err)
		}
		connPool.ReleaseConn(conn)
	}
	_, err := connPool.GetConn()
	if err != nil {
		t.Errorf("test failed, expected nil, got: %s", err)
	}
}
