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
	tables := []struct {
		input    string
		expected []byte
	}{
		{"+OK\r\n", []byte("OK")},
		{"-Error message\r\n", []byte("Error message")},
		{":1000\r\n", []byte("1000")},
		{"$6\r\nfoobar\r\n", []byte("foobar")},
		{"*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+Foo\r\n-Bar\r\n", []byte("")},
		{"*5\r\n:1\r\n:2\r\n:3\r\n:4\r\n$6\r\nfoobar\r\n", []byte("")},
		{"*0\r\n", []byte("")},
		{"*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n", []byte("")},
		{"*3\r\n:1\r\n:2\r\n:3\r\n", []byte("")},
		{"*-1\r\n", []byte("")},
	}
	for _, table := range tables {
		str := table.input
		b := bytes.NewBufferString(str)
		respReader := NewRESPReader(b)
		decoded := respReader.ReadResp()
		fmt.Printf("decoded: %c\n", decoded)
		//expected := table.expected
		//		if decoded != expected {
		//			t.Errorf("test failed, expected: %s, got: %s", table.expected, decoded)
		//		}
	}
}
