package main

import (
	"bytes"
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
		fmt.Fprintf(conn, encodeCommandString(commandStr))

		reply := make([]byte, 8192)

		_, err = conn.Read(reply)
		if err != nil {
			fmt.Println("Read server reply failed:", err.Error())
			os.Exit(1)
		}

		fmt.Println("reply from server: ", decodeResp(reply))
	}

	conn.Close()
}

// translate command string into strings which the server can understand
func encodeCommandString(commandString string) string {
	trimmed := strings.Trim(commandString, " \n")
	//fmt.Printf("trimmed: %s", trimmed)
	commandSlice := strings.Split(trimmed, " ")
	return makeArray(commandSlice)
}

func decodeResp(s []byte) string {
	switch s[0] {
	case '+':
		resp, _ := decodeSimpleString(s[1:], 0)
		return resp
	case '-':
		resp, _ := decodeErrorString(s[1:], 0)
		return resp
	case ':':
		resp, _ := decodeIntString(s[1:], 0)
		return resp
	case '$':
		resp, _ := decodeBulkString(s[1:], 0)
		return resp
	case '*':
		resp, _ := decodeArrayString(s[1:], 0)
		return fmt.Sprint(resp)
	default:
		return "not gonna happen"
	}
}

// make array string
func makeArray(strSlice []string) string {
	sliceLen := len(strSlice)
	// strings.Builder was just added on go 1.10(which was released on 2018-02-16)
	//var b strings.Builder
	var b bytes.Buffer
	fmt.Fprintf(&b, "*%d\r\n", sliceLen)
	for _, str := range strSlice {
		b.WriteString(makeBulkString(str))
	}
	return b.String()
}

// make bulk string
func makeBulkString(str string) string {
	strLen := len(str)
	//var b strings.Builder
	var b bytes.Buffer
	fmt.Fprintf(&b, "$%d\r\n%s\r\n", strLen, str)
	return b.String()
}

func decodeSimpleString(s []byte, index int) (string, int) {
	var b bytes.Buffer
	var c byte
	var i int
	for i = index; ; i++ {
		c = s[i]
		if c != '\r' {
			b.WriteByte(c)
		} else {
			break
		}
	}
	return b.String(), i + 2
}

func decodeErrorString(s []byte, index int) (string, int) {
	return decodeSimpleString(s, index)
}

func decodeIntString(s []byte, index int) (string, int) {
	return decodeSimpleString(s, index)
}

func decodeBulkString(s []byte, index int) (string, int) {
	var b bytes.Buffer
	var i int
	lenStr, strIndex := decodeSimpleString(s, index)
	strLen, _ := strconv.Atoi(lenStr)
	if strLen == -1 {
		return "nil", strIndex
	}
	for i = strIndex; i < strLen+strIndex; i++ {
		b.WriteByte(s[i])
	}
	return b.String(), i + 2
}

func decodeArrayString(s []byte, index int) ([]string, int) {
	var result []string
	lenStr, strIndex := decodeSimpleString(s, index)
	arrayLen, _ := strconv.Atoi(lenStr)
	if arrayLen == -1 {
		return []string{"nil"}, strIndex
	}
	var arrayElement string
	for i := 0; i < arrayLen; i++ {
		mark := s[strIndex]
		switch mark {
		case '+':
			arrayElement, strIndex = decodeSimpleString(s, strIndex+1)
		case '-':
			arrayElement, strIndex = decodeErrorString(s, strIndex+1)
		case ':':
			arrayElement, strIndex = decodeIntString(s, strIndex+1)
		case '$':
			arrayElement, strIndex = decodeBulkString(s, strIndex+1)
		}
		result = append(result, arrayElement)
	}
	return result, strIndex
}
