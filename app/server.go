package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	buf := make([]byte, 4096)

	// 계속해서 명령어 읽기
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			return
		}
		if n == 0 {
			continue
		}

		data := buf[:n]

		// RESP 프로토콜 형식인지 확인
		if bytes.HasPrefix(data, []byte("*")) {
			// RESP 프로토콜: 하나의 PONG 응답
			conn.Write([]byte("+PONG\r\n"))
		} else {
			// 일반 텍스트: 줄 단위로 처리
			commands := bytes.Split(data, []byte("\n"))
			for _, cmd := range commands {
				if len(bytes.TrimSpace(cmd)) > 0 {
					conn.Write([]byte("+PONG\r\n"))
				}
			}
		}
	}
}
