package main

import (
	"fmt"
	"log"
	"net"
	"runtime"
	"strconv"
)

func CheckError(err error) {
	if err != nil {
		b := make([]byte, 2048)
		n := runtime.Stack(b, false)
		stack := string(b[:n])
		log.Println("stack", stack)
		log.Fatal("Error: ", err)
	}
}

func SendClientMessagesToUdpServer() {
	ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	CheckError(err)

	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	CheckError(err)

	Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	CheckError(err)

	defer Conn.Close()
	i := 0
	for {
		msg := strconv.Itoa(i)
		i++
		buf := []byte(msg)
		_, err := Conn.Write(buf)
		log.Println("sending msg: " + msg)
		if err != nil {
			fmt.Println(msg, err)
		}
		if i > 10 {
			break
		}
	}
}
