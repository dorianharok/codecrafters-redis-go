package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
)

var (
	redisMap = make(map[string]string)
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
					redisMap[key] = value
					conn.Write([]byte("+OK\r\n"))
				}
			case "GET":
				if len(parts) >= 5 {
					key := string(parts[4])
					value, ok := redisMap[key]
					if ok {
						conn.Write([]byte("+" + value + "\r\n"))
					}
				}
			}
		}
	}
}
