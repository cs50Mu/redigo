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

func TestReadResp(t *testing.T) {
	// simple str
	b := bytes.NewBufferString("+OK\r\n")
	respReader := NewRESPReader(b)
	reply, _ := respReader.ReadResp()
	expected := "OK"
	if string(reply.stringVal) != expected {
		t.Errorf("test failed, expected: %s, got: %s", expected, string(reply.stringVal))
	}

	// error str
	b = bytes.NewBufferString("-Error message\r\n")
	respReader = NewRESPReader(b)
	reply, err := respReader.ReadResp()
	expected = "Error message"
	if err.Error() != expected {
		t.Errorf("test failed, expected: %s, got: %s", expected, err.Error())
	}

	// integer str
	b = bytes.NewBufferString(":1000\r\n")
	respReader = NewRESPReader(b)
	reply, err = respReader.ReadResp()
	var intExpected = int64(1000)
	if reply.integerVal != intExpected {
		t.Errorf("test failed, expected: %d, got: %d", intExpected, reply.integerVal)
	}

	// bulk str
	b = bytes.NewBufferString("$6\r\nfoobar\r\n")
	respReader = NewRESPReader(b)
	reply, err = respReader.ReadResp()
	expected = "foobar"
	if string(reply.stringVal) != expected {
		t.Errorf("test failed, expected: %s, got: %s", expected, string(reply.stringVal))
	}

	// array str
	b = bytes.NewBufferString("*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n")
	respReader = NewRESPReader(b)
	reply, err = respReader.ReadResp()
	arrayReply := reply.arrayVal
	if string(arrayReply[0].stringVal) != "foo" && string(arrayReply[1].stringVal) != "bar" {
		t.Errorf("test failed, expected: [foo, bar], got: [%s, %s]", arrayReply[0].stringVal, arrayReply[1].stringVal)
	}

	// empty str
	b = bytes.NewBufferString("$0\r\n\r\n")
	respReader = NewRESPReader(b)
	reply, err = respReader.ReadResp()
	expected = ""
	if string(reply.stringVal) != expected {
		t.Errorf("test failed, expected: %s, got: %s", expected, string(reply.stringVal))
	}
	// emtpy array
	b = bytes.NewBufferString("*0\r\n")
	respReader = NewRESPReader(b)
	reply, err = respReader.ReadResp()
	if len(reply.arrayVal) != 0 {
		t.Errorf("test failed, expected: empty array, got: %d element in array", len(reply.arrayVal))
	}
	// nil object as bulk str
	b = bytes.NewBufferString("$-1\r\n")
	respReader = NewRESPReader(b)
	reply, err = respReader.ReadResp()
	if reply.stringVal != nil {
		t.Errorf("test failed, expected: nil, got: %s", string(reply.stringVal))
	}
	// nil object as array str
	b = bytes.NewBufferString("*-1\r\n")
	respReader = NewRESPReader(b)
	reply, err = respReader.ReadResp()
	if reply.arrayVal != nil {
		t.Errorf("test failed, expected: nil")
	}
}

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
	if len(res) != 0 {
		t.Errorf("test failed, expected: empty array, got: %s", res)
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

func TestPipeline(t *testing.T) {
	fmt.Printf("testing pipeline\n")
	client, _ := NewRedisClient("127.0.0.1", "6379")
	pipeline, _ := client.Pipeline()
	pipeline.AddCommand("set", "x", "1")
	pipeline.AddCommand("incr", "x")
	pipeline.AddCommand("get", "x")
	pipeline.AddCommand("info")
	r, err := pipeline.Exec()
	if err != nil {
		fmt.Printf("got error while exec: %s\n", err)
	}
	fmt.Printf("reply from set: %s\n", r[0].stringVal)
	fmt.Printf("reply from incr: %d\n", r[1].integerVal)
	fmt.Printf("reply from get: %s\n", r[2].stringVal)
	fmt.Printf("reply from info: %s\n", r[3].stringVal)
	pipeline.Close()
}

func TestTransaction(t *testing.T) {
	fmt.Printf("testing transaction\n")
	client, _ := NewRedisClient("127.0.0.1", "6379")
	tx, _ := client.Transaction()
	tx.Watch("x")
	tx.AddCommand("set", "x", "1")
	tx.AddCommand("incr", "x")
	tx.AddCommand("get", "x")
	r, err := tx.Exec()
	if err != nil {
		panic(err)
	}
	fmt.Printf("reply from set: %s\n", r[0].stringVal)
	fmt.Printf("reply from incr: %d\n", r[1].integerVal)
	fmt.Printf("reply from get: %s\n", r[2].stringVal)
	tx.Close()
}

func TestPubSub(t *testing.T) {
	fmt.Printf("\ntesting pubsub\n")
	client2, _ := NewRedisClient("127.0.0.1", "6379")
	ps2, _ := client2.PubSub()
	ps2.Subscribe("test_channel")
	client3, _ := NewRedisClient("127.0.0.1", "6379")
	ps3, _ := client3.PubSub()
	ps3.Publish("test_channel", "hello")
	ps2.Unsubscribe("test_channel")
	for i := 0; i < 3; i++ {
		message, _ := ps2.Receive()
		fmt.Printf("message: %s\n", message)
	}
	ps3.Close()
	ps2.Close()
}

func TestScripting(t *testing.T) {
	fmt.Printf("testing scripting\n")
	client, _ := NewRedisClient("127.0.0.1", "6379")
	res, _ := client.ScriptLoad(`return redis.call('get','foo')`)
	fmt.Printf("reply from script load: %s\n", res)
}
