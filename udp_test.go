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
		clientPort   = "10101"
	)
	go outgoingServer(outgoingPort, output)

	go clientSend(clientPort, outgoingPort, "CONNECT 1")
	<-output
	if Clients[Id("1")] == nil {
		t.Error("CONNECT not working")
	}
	go clientSend(clientPort, outgoingPort, "DISCONNECT 1")
	<-output
	if Clients[Id("1")] != nil {
		t.Error("DISCONNECT not working")
	}
}

func Test_ClientMustBeRemovedTimeout(t *testing.T) {
	var (
		clientPort   = "10101"
		outgoingPort = "10002"
		output       = make(chan string)
	)
	go outgoingServer(outgoingPort, output)
	DISCONNECT_TIMEOUT = time.Millisecond * 10
	clientSend(clientPort, outgoingPort, "CONNECT 2")
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
		clientPort   = "10103"
	)
	go outgoingServer(outgoingPort, output)
	DISCONNECT_TIMEOUT = time.Millisecond * 10
	clientSend(clientPort, outgoingPort, "CONNECT 3")
	msg := <-output
	log.Println(msg)
	sleep := DISCONNECT_TIMEOUT / 3
	for i := 1; i <= 5; i++ {
		log.Println("sleeping for ", sleep, "in test")
		time.Sleep(sleep)
		clientSend(clientPort, outgoingPort, "ALIVE 3")
	}

	if Clients[Id("3")] == nil {
		t.Error("ALIVE command not working")
	}
}

func clientSend(clientPort, serverPort string, msg string) {
	ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+serverPort)
	CheckError(err)

	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+clientPort)
	CheckError(err)

	Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	CheckError(err)

	defer Conn.Close()
	log.Println("sending in test in clientSend ", msg)
	buf := []byte(msg)
	_, err = Conn.Write(buf)
	CheckError(err)
}

func TestStreamerWithOneClient(t *testing.T) {
	const (
		incomingPort = "10020"
		outgoingPort = "10011"
		clientPort   = "10012"

		masterPort = "10555"
	)
	output := make(chan string)

	go RunStreamer(incomingPort, outgoingPort, output)
	log.Println(<-output)
	clientSend(clientPort, outgoingPort, "CONNECT 33")
	log.Println(<-output)

	if Clients[Id("33")] == nil {
		t.Error("client hasn't connected")
		t.FailNow()
	}
	log.Println("len(Clients)")
	log.Println(len(Clients))
	log.Println("client connected")

	clientOutput := make(chan string)
	go clientListen(clientPort, clientOutput)
	log.Println(<-clientOutput)
	msg := "hello clients"
	clientSend(masterPort, incomingPort, msg)
	if actual := <-clientOutput; actual != msg {
		t.Errorf("client didn't receive message '%s' from Streamer, instead it received '%s'", msg, actual)
	}
}

func TestStreamerWithOneClientWithoutConnecting(t *testing.T) {
	const (
		incomingPort = "10020"
		outgoingPort = "10011"
		clientPort   = "10017"

		masterPort = "10555"
	)
	output := make(chan string)

	go RunStreamer(incomingPort, outgoingPort, output)
	log.Println(<-output)

	addr, err := net.ResolveUDPAddr("udp", ":"+clientPort)
	CheckError(err)
	Clients[Id("33")] = &Client{
		Addr:          addr,
		Id:            Id("33"),
		LastAliveTime: time.Now(),
	}

	log.Println("len(Clients)")
	log.Println(len(Clients))
	log.Println("client connected")

	clientOutput := make(chan string)
	go clientListen(clientPort, clientOutput)
	log.Println(<-clientOutput)
	msg := "hello clients"
	clientSend(masterPort, incomingPort, msg)
	if actual := <-clientOutput; actual != msg {
		t.Errorf("client didn't receive message '%s' from Streamer, instead it received '%s'", msg, actual)
	}
}

func clientListen(clientPort string, output chan string) {
	/* Lets prepare a address at any address at port */
	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+clientPort)
	CheckError(err)

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()

	buf := make([]byte, 1024)

	for {
		output <- "started listening"
		n, _, err := ServerConn.ReadFromUDP(buf)
		log.Println("evrikaevrikaevrikaevrikaevrikaevrika")

		CheckError(err)

		output <- string(buf[0:n])
	}
}

func Test_sendTo(t *testing.T) {
	outgoingPort := "10188"
	clientPort := "10189"
	addr, err := net.ResolveUDPAddr("udp", ":"+clientPort)
	CheckError(err)

	client := &Client{
		Addr: addr,
	}
	output := make(chan string)
	go clientListen(clientPort, output)
	log.Println(<-output)
	msg := "hello client in test"
	sendTo(client, outgoingPort, []byte(msg))
	if actual := <-output; actual != msg {
		log.Println(actual)
		t.Error("sendTo failed")
	}
}
