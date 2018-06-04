package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strconv"
)

// Reply from redis server
type Reply struct {
	// 为了放nil才使用[]byte类型，否则用string是最合适的
	stringVal  []byte
	integerVal int64
	arrayVal   []*Reply
}

const (
	arrayStrPrefix   byte   = '*'
	bulkStrPrefix    byte   = '$'
	simpleStrPrefix  byte   = '+'
	integerStrPrefix byte   = ':'
	errorStrPrefix   byte   = '-'
	terminator       string = "\r\n"
)

// RESPWriter encodes command into
// language that server can understand
type RESPWriter struct {
	*bufio.Writer
}

// NewRESPWriter returns a RESPWriter
func NewRESPWriter(w io.Writer) *RESPWriter {
	return &RESPWriter{
		Writer: bufio.NewWriter(w),
	}
}

// WriteCommand write command to buffer
func (w *RESPWriter) WriteCommand(commands ...string) error {
	commandLen := len(commands)
	w.WriteByte(arrayStrPrefix)
	w.WriteString(strconv.Itoa(commandLen))
	w.WriteString(terminator)

	for _, c := range commands {
		w.writeBulkString(c)
	}
	return w.Flush()
}

func (w *RESPWriter) writeBulkString(str string) {
	strLen := len(str)
	w.WriteByte(bulkStrPrefix)
	w.WriteString(strconv.Itoa(strLen))
	w.WriteString(terminator)
	w.WriteString(str)
	w.WriteString(terminator)
}

// RESPReader decodes command
// from redis server
type RESPReader struct {
	*bufio.Reader
}

// NewRESPReader returns a RESPReader
func NewRESPReader(r io.Reader) *RESPReader {
	return &RESPReader{
		Reader: bufio.NewReader(r),
	}
}

// ReadResp read resp
func (r *RESPReader) ReadResp() (*Reply, error) {
	prefix, _ := r.ReadByte()
	switch prefix {
	case simpleStrPrefix:
		return r.readSimpleStr()
	case integerStrPrefix:
		reply, _ := r.readSimpleStr()
		intVal, _ := strconv.ParseInt(string(reply.stringVal), 10, 64)
		return &Reply{integerVal: intVal}, nil
	case errorStrPrefix:
		reply, _ := r.readSimpleStr()
		return nil, errors.New(string(reply.stringVal))
	case bulkStrPrefix:
		//		var b bytes.Buffer
		return r.readBulkStr()
	case arrayStrPrefix:
		return r.readArrayStr()
	default:
		return nil, errors.New("not supported data type")
	}
}

func (r *RESPReader) readSimpleStr() (*Reply, error) {
	str := r.readLine()
	return &Reply{stringVal: str}, nil
}

func (r *RESPReader) readLine() []byte {
	var b bytes.Buffer
	var c byte
	for {
		c, _ = r.ReadByte()
		if c != '\r' {
			b.WriteByte(c)
		} else {
			break
		}
	}
	// consume '\n'
	r.ReadByte()
	return b.Bytes()
}

func (r *RESPReader) readBulkStr() (*Reply, error) {
	var b bytes.Buffer
	lenStr := string(r.readLine())
	strLen, _ := strconv.Atoi(lenStr)
	if strLen == -1 {
		return &Reply{stringVal: nil}, nil
	}
	for i := strLen; i > 0; i-- {
		c, _ := r.ReadByte()
		b.WriteByte(c)
	}
	// consume '\r\n'
	r.ReadByte()
	r.ReadByte()
	return &Reply{stringVal: b.Bytes()}, nil
}

func (r *RESPReader) readArrayStr() (*Reply, error) {
	s := make([]*Reply, 0)
	lenStr := string(r.readLine())
	strLen, _ := strconv.Atoi(lenStr)
	if strLen == 0 {
		return &Reply{arrayVal: nil}, nil
	}
	for i := strLen; i > 0; i-- {
		res, err := r.ReadResp()
		if err != nil {
			return nil, err
		}
		s = append(s, res)
	}
	return &Reply{arrayVal: s}, nil
}
