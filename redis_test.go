package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestWriteCommand(t *testing.T) {
	tables := []struct {
		input    []string
		expected string
	}{
		{[]string{"keys", "*"}, "*2\r\n$4\r\nkeys\r\n$1\r\n*\r\n"},
		{[]string{"foo", "bar"}, "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"},
		//{[]string{"expire", "x", "20"}, ""},
	}

	for _, table := range tables {
		var b bytes.Buffer
		respWriter := NewRESPWriter(&b)
		respWriter.WriteCommand(table.input...)
		encoded := b.String()
		if encoded != table.expected {
			t.Errorf("test failed, expected: %s, got: %s", table.expected, encoded)
		}
	}
}

//func TestReadResp(t *testing.T) {
//	tables := []struct {
//		input    string
//		expected []byte
//	}{
//		{"+OK\r\n", []byte("OK")},
//		{"-Error message\r\n", []byte("Error message")},
//		{":1000\r\n", []byte("1000")},
//		{"$6\r\nfoobar\r\n", []byte("foobar")},
//		{"*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+Foo\r\n-Bar\r\n", []byte("")},
//		{"*5\r\n:1\r\n:2\r\n:3\r\n:4\r\n$6\r\nfoobar\r\n", []byte("")},
//		{"*0\r\n", []byte("")},
//		{"*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n", []byte("")},
//		{"*3\r\n:1\r\n:2\r\n:3\r\n", []byte("")},
//		{"*-1\r\n", []byte("")},
//	}
//	for _, table := range tables {
//		str := table.input
//		b := bytes.NewBufferString(str)
//		respReader := NewRESPReader(b)
//		decoded := respReader.ReadResp()
//		fmt.Printf("decoded: %c\n", decoded)
//		//expected := table.expected
//		//		if decoded != expected {
//		//			t.Errorf("test failed, expected: %s, got: %s", table.expected, decoded)
//		//		}
//	}
//}

func TestGet(t *testing.T) {
	client, _ := NewRedisClient("127.0.0.1", "6379")
	expected := "hello linuxfish"
	ok, err := client.Set("x", expected)
	if err != nil {
		fmt.Println(err)
		return
	}
	if ok {
		fmt.Println("set ok")
	} else {
		fmt.Println("set failed")
	}
	val, _ := client.Get("x")
	if string(val) != expected {
		t.Errorf("test failed, expected: %s, got: %s", expected, val)
	}
}

// get not exist key returns nil
func TestGetNotExist(t *testing.T) {
	client, _ := NewRedisClient("127.0.0.1", "6379")
	val, _ := client.Get("not_exist")
	if val != nil {
		t.Errorf("test failed, expected: nil, got: %s", val)
	}
}

func TestExpire(t *testing.T) {
	client, _ := NewRedisClient("127.0.0.1", "6379")
	client.Set("x", "test expire")
	_, err := client.Expire("x", 10)
	if err != nil {
		t.Errorf("expire error: %s", err)
	}
	ttl, _ := client.TTL("x")
	fmt.Printf("ttl: %d\n", ttl)
	if ttl < 0 {
		t.Errorf("test failed, %d should greater than zero", ttl)
	}
}

func TestKeys(t *testing.T) {
	client, _ := NewRedisClient("127.0.0.1", "6379")
	res, _ := client.Keys("*")
	fmt.Println(res)
	client.Select(1)
	res, _ = client.Keys("*")
	if res != nil {
		t.Errorf("test failed, expected: nil, got: %s", res)
	}
}

func TestMget(t *testing.T) {
	client, _ := NewRedisClient("127.0.0.1", "6379")
	kvs := map[string]string{
		"name":   "linuxfish",
		"gender": "male",
	}
	client.Mset(kvs)
	vals, _ := client.Mget("name", "not_exist", "gender")
	fmt.Printf("%s\n", vals)
	if vals[1] != nil {
		t.Errorf("test failed, expected: nil, got: %s", vals[1])
	}
}

func TestIncr(t *testing.T) {
	client, _ := NewRedisClient("127.0.0.1", "6379")
	client.Del("progress")
	res, _ := client.Incr("progress")
	if res != int64(1) {
		t.Errorf("test failed, expected: 1, got: %d", res)
	}
}

func TestIncrByFloat(t *testing.T) {
	client, _ := NewRedisClient("127.0.0.1", "6379")
	client.Del("float")
	client.Set("float", "0.1")
	res, _ := client.IncrByFloat("float", float64(0.1))
	fmt.Printf("incred: %f\n", res)
}

func TestScan(t *testing.T) {
	client, _ := NewRedisClient("127.0.0.1", "6379")
	nextCursor, keys, _ := client.Scan(0, "", 3)
	fmt.Printf("keys: %s\n", keys)
	nextCursor, keys, _ = client.Scan(nextCursor, "", 3)
	fmt.Printf("keys: %s\n", keys)
}
