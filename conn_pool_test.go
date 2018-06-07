package main

import (
	"fmt"
	"testing"
)

func TestPush(t *testing.T) {
	fmt.Println("testing conn_pool")
	s := newStack()
	var c Conn
	s.push(c)
	if s.length() != 1 {
		t.Errorf("test failed, expected: 1, got: %d", s.length())
	}
}

func TestConnPool(t *testing.T) {
	connPool := NewConnPool("127.0.0.1", "6379", 10)
	var conns []Conn
	for i := 1; i <= 10; i++ {
		conn, err := connPool.GetConn()
		if err != nil {
			fmt.Printf("got error: %s\n", err)
		}
		conns = append(conns, conn)
	}
	_, err := connPool.GetConn()
	if err == nil {
		t.Errorf("test failed, expected not nil, got: nil")
	}

	for _, c := range conns {
		connPool.ReleaseConn(c)
	}
	_, err = connPool.GetConn()
	if err != nil {
		t.Errorf("test failed, expected nil, got: %s", err)
	}
}
