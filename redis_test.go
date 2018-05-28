package main

import (
	"reflect"
	"testing"
)

func TestEncodeCommandString(t *testing.T) {
	commandStr := "keys *"
	encodedStr := encodeCommandString(commandStr)
	want := "*2\r\n$4\r\nkeys\r\n$1\r\n*\r\n"
	if encodedStr != want {
		t.Errorf("test failed, want: %s, got: %s", want, encodedStr)
	}
}

func TestDecodeSimpleStr(t *testing.T) {
	str := "OK\r\n"
	decoded, _ := decodeSimpleString([]byte(str), 0)
	want := "OK"
	if decoded != want {
		t.Errorf("test failed, want: %s, got: %s", want, decoded)
	}
}

func TestDecodeErrorString(t *testing.T) {
	str := "ERR unknown command 'foobar'\r\n"
	decoded, _ := decodeErrorString([]byte(str), 0)
	want := "ERR unknown command 'foobar'"
	if decoded != want {
		t.Errorf("test failed, want: %s, got: %s", want, decoded)
	}
}

func TestDecodeBulkString(t *testing.T) {
	str := "6\r\nfoobar\r\n"
	decoded, _ := decodeBulkString([]byte(str), 0)
	want := "foobar"
	if decoded != want {
		t.Errorf("test failed, want: %s, got: %s", want, decoded)
	}
}

func TestDecodeArrayString(t *testing.T) {
	str := "2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"
	decoded, _ := decodeArrayString([]byte(str), 0)
	want := []string{"foo", "bar"}
	t.Log(decoded)
	for i, v := range decoded {
		t.Logf("%d: %s", i, v)
	}
	if !reflect.DeepEqual(want, decoded) {
		t.Errorf("test failed, want: %s, got: %s", want, decoded)
	}

	str = "5\r\n:1\r\n:2\r\n:3\r\n:4\r\n$6\r\nfoobar\r\n"
	decoded, _ = decodeArrayString([]byte(str), 0)
	want = []string{"1", "2", "3", "4", "foobar"}
	if !reflect.DeepEqual(want, decoded) {
		t.Errorf("test failed, want: %s, got: %s", want, decoded)
	}
}

func TestNullBulkString(t *testing.T) {
	str := "-1\r\n"
	decoded, _ := decodeBulkString([]byte(str), 0)
	want := "nil"
	if decoded != want {
		t.Errorf("test failed, want: %s, got: %s", want, decoded)
	}
}

func TestNullElementInArray(t *testing.T) {
	str := "3\r\n$3\r\nfoo\r\n$-1\r\n$3\r\nbar\r\n"
	decoded, _ := decodeArrayString([]byte(str), 0)
	want := []string{"foo", "nil", "bar"}
	if !reflect.DeepEqual(want, decoded) {
		t.Errorf("test failed, want: %s, got: %s", want, decoded)
	}
}
