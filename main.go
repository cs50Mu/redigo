package main

import (
	"bytes"
	"io"
	//"flag"
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	//commandPtr := flag.String("c", "", "redis command")

	//flag.Parse()

	conn, err := net.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		// handle error
		fmt.Printf("connect server error: %v", err)
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("redis> ")
		commandStr, _ := reader.ReadString('\n')
		//commandStr := *commandPtr
		trimmed := strings.Trim(commandStr, " \n")
		//fmt.Printf("trimmed: %s", trimmed)
		commandSlice := strings.Split(trimmed, " ")
		respWriter := NewRESPWriter(conn)
		respWriter.WriteCommand(commandSlice...)
		//fmt.Fprintf(conn, encodeCommandString(commandStr))

		reply := make([]byte, 8192)

		_, err = conn.Read(reply)
		if err != nil {
			fmt.Println("Read server reply failed:", err.Error())
			os.Exit(1)
		}

		respReader := NewRESPReader(bytes.NewBuffer(reply))
		fmt.Println("reply from server: ", string(respReader.ReadResp()))
	}

	conn.Close()
}

const (
	arrayStrPrefix string = "*"
	bulkStrPrefix  string = "$"
	terminator     string = "\r\n"
)

// RESPWriter encodes command into
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
	w.WriteString(arrayStrPrefix)
	w.WriteString(strconv.Itoa(commandLen))
	w.WriteString(terminator)

	for _, c := range commands {
		w.writeBulkString(c)
	}
	return w.Flush()
}

func (w *RESPWriter) writeBulkString(str string) {
	strLen := len(str)
	w.WriteString(bulkStrPrefix)
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
func (r *RESPReader) ReadResp() []byte {
	prefix, _ := r.ReadByte()
	switch prefix {
	case '+', '-', ':':
		return r.readSimpleStr()
	case '$':
		//		var b bytes.Buffer
		return r.readBulkStr()
	case '*':
		return r.readArrayStr()
	default:
		return []byte{}
	}
}

func (r *RESPReader) readSimpleStr() []byte {
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

func (r *RESPReader) readBulkStr() []byte {
	var b bytes.Buffer
	lenStr := string(r.readSimpleStr())
	strLen, _ := strconv.Atoi(lenStr)
	if strLen == -1 {
		return []byte("nil")
	}
	for i := strLen; i > 0; i-- {
		c, _ := r.ReadByte()
		b.WriteByte(c)
	}
	// consume '\r\n'
	r.ReadByte()
	r.ReadByte()
	return b.Bytes()
}

func (r *RESPReader) readArrayStr() []byte {
	s := make([]byte, 0)
	lenStr := string(r.readSimpleStr())
	strLen, _ := strconv.Atoi(lenStr)
	if strLen == -1 {
		return []byte("nil")
	}
	for i := strLen; i > 0; i-- {
		res := r.ReadResp()
		s = append(s, res...)
	}
	return s
}
