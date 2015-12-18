package main

import (
	"log"
	"net"
	"testing"
	"time"
)

func Test_ConnectDisconnectClient(t *testing.T) {
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
	DISCONNECT_TIMEOUT = time.Millisecond * 10
	ClientSendToUdpServer(outgoingPort, "CONNECT 3")
	msg := <-output
	log.Println(msg)
	for i := 1; i <= 10; i++ {
		time.Sleep(DISCONNECT_TIMEOUT / 3)
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
