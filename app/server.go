package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// 값과 만료시간을 함께 저장할 구조체
type RedisValue struct {
	value    string
	expireAt *time.Time // nil이면 만료시간 없음
}

var (
	redisMap = make(map[string]RedisValue)
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			return
		}

		data := buf[:n]
		parts := bytes.Split(data, []byte("\r\n"))

		if len(parts) >= 4 {
			cmd := strings.ToUpper(string(parts[2]))

			switch cmd {
			case "PING":
				conn.Write([]byte("+PONG\r\n"))
			case "ECHO":
				if len(parts) >= 5 {
					arg := string(parts[4])
					response := fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)
					conn.Write([]byte(response))
				}
			case "SET":
				if len(parts) >= 7 {
					key := string(parts[4])
					value := string(parts[6])

					var expireAt *time.Time
					if len(parts) >= 9 && strings.ToUpper(string(parts[8])) == "PX" && len(parts) >= 11 {
						ms, err := strconv.ParseInt(string(parts[10]), 10, 64)
						if err == nil {
							t := time.Now().Add(time.Duration(ms) * time.Millisecond)
							expireAt = &t

							go func() {
								time.Sleep(time.Duration(ms) * time.Millisecond)
								delete(redisMap, key)
							}()
						}
					}

					redisMap[key] = RedisValue{
						value:    value,
						expireAt: expireAt,
					}
					conn.Write([]byte("+OK\r\n"))
				}
			case "GET":
				if len(parts) >= 5 {
					key := string(parts[4])
					if val, exists := redisMap[key]; exists {
						if val.expireAt == nil || time.Now().Before(*val.expireAt) {
							response := fmt.Sprintf("$%d\r\n%s\r\n", len(val.value), val.value)
							conn.Write([]byte(response))
						} else {
							delete(redisMap, key)
							conn.Write([]byte("$-1\r\n"))
						}
					} else {
						conn.Write([]byte("$-1\r\n"))
					}
				}
			}
		}
	}
}
