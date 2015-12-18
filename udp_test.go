package main

import (
	"log"
	"net"
	"testing"
	"time"
)

func Test_ConnectDisconnectClient(t *testing.T) {
	var (
		outgoingPort = "10001"
		output       = make(chan string)
	)
	go outgoingServer(outgoingPort, output)

	go ClientSendToUdpServer(outgoingPort, "CONNECT 1")
	<-output
	if Clients[Id("1")] == nil {
		t.Error("CONNECT not working")
	}
	go ClientSendToUdpServer(outgoingPort, "DISCONNECT 1")
	<-output
	if Clients[Id("1")] != nil {
		t.Error("DISCONNECT not working")
	}
}

func Test_ClientMustBeRemovedTimeout(t *testing.T) {
	var (
		outgoingPort = "10002"
		output       = make(chan string)
	)
	go outgoingServer(outgoingPort, output)
	DISCONNECT_TIMEOUT = time.Millisecond * 10
	ClientSendToUdpServer(outgoingPort, "CONNECT 2")
	msg := <-output
	log.Println(msg)
	time.Sleep(3 * DISCONNECT_TIMEOUT)
	if Clients[Id("2")] != nil {
		t.Error("client must be disconnected because of timeout")
	}
}

func Test_KeepAlive(t *testing.T) {
	var (
		outgoingPort = "10003"
		output       = make(chan string)
	)
	go outgoingServer(outgoingPort, output)
	DISCONNECT_TIMEOUT = time.Millisecond * 10
	ClientSendToUdpServer(outgoingPort, "CONNECT 3")
	msg := <-output
	log.Println(msg)
	sleep := DISCONNECT_TIMEOUT / 3
	for i := 1; i <= 5; i++ {
		log.Println("sleeping for ", sleep, "in test")
		time.Sleep(sleep)
		ClientSendToUdpServer(outgoingPort, "ALIVE 3")
	}

	if Clients[Id("3")] == nil {
		t.Error("ALIVE command not working")
	}
}

func ClientSendToUdpServer(serverPort string, msg string) {
	ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+serverPort)
	CheckError(err)

	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	CheckError(err)

	Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	CheckError(err)

	defer Conn.Close()
	buf := []byte(msg)
	_, err = Conn.Write(buf)
	CheckError(err)
}
