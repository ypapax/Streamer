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
	log.Println(<-output)
	clientSend(clientPort, outgoingPort, "CONNECT 1")
	log.Println("hello")
	log.Println(<-output)
	log.Println("hello2")

	if Clients[Id("1")] == nil {
		t.Error("CONNECT not working")
		t.FailNow()
	}
	log.Println("CONNECT is working")
	clientSend(clientPort, outgoingPort, "DISCONNECT 1")
	log.Println(<-output)
	if Clients[Id("1")] != nil {
		t.Error("DISCONNECT not working")
		t.FailNow()
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
	log.Println(<-output)
	DISCONNECT_TIMEOUT = time.Millisecond * 100
	clientSend(clientPort, outgoingPort, "CONNECT 3")
	log.Println(<-output)
	sleep := DISCONNECT_TIMEOUT / 3
	for i := 1; i <= 5; i++ {
		log.Println("len(Clients)", len(Clients))
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
	sendStr(clientPort, ServerAddr, msg)
}

func TestStreamerWithOneClient(t *testing.T) {
	const (
		incomingPort = "30020"
		outgoingPort = "30011"
		clientPort   = "30012"

		masterPort = "30555"
	)
	output := make(chan string)

	go RunStreamer(incomingPort, outgoingPort, output)
	log.Println(<-output)
	clientSend(clientPort, outgoingPort, "CONNECT 33")
	log.Println(<-output)
	log.Println(<-output)

	if Clients[Id("33")] == nil {
		t.Error("client hasn't connected")
		t.FailNow()
	}
	log.Println("len(Clients)")
	log.Println(len(Clients))
	log.Println("client connected")

	data := make(chan *Data)
	go listen(clientPort, output, data)
	msg := "hello clients"
	clientSend(masterPort, incomingPort, msg)
	log.Println(<-output)
	// log.Println(<-output)
	d := <-data
	if actual := d.String(); actual != msg {
		t.Errorf("client didn't receive message '%s' from Streamer, instead it received '%s'", msg, actual)
	}
}

func TestStreamerWithOneClientWithoutConnecting(t *testing.T) {
	const (
		incomingPort = "20020"
		outgoingPort = "20011"
		clientPort   = "20017"

		masterPort = "20555"
	)
	logger := make(chan string)

	go RunStreamer(incomingPort, outgoingPort, logger)

	log.Println(<-logger)

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

	dataChan := make(chan *Data)
	go listen(clientPort, logger, dataChan)
	log.Println(<-logger)
	log.Println(<-logger)
	// log.Println(<-logger)
	msg := "hello clients"
	clientSend(masterPort, incomingPort, msg)
	data := <-dataChan
	if actual := data.String(); actual != msg {
		t.Errorf("client didn't receive message '%s' from Streamer, instead it received '%s'", msg, actual)
	}
}

func TestIncomingServer(t *testing.T) {
	const (
		incomingPort = "10020"
		clientPort   = "10017"

		masterPort = "10555"
	)
	logger := make(chan string)

	go incomingServer(incomingPort, logger)

	log.Println(<-logger)

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

	dataChan := make(chan *Data)
	logger2 := make(chan string)
	go listen(clientPort, logger2, dataChan)
	log.Println(<-logger2)
	msg := "hello clients"
	clientSend(masterPort, incomingPort, msg)
	data := <-dataChan
	if actual := data.String(); actual != msg {
		t.Errorf("client didn't receive message '%s' from Streamer, instead it received '%s'", msg, actual)
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
	logger := make(chan string)
	dataChan := make(chan *Data)
	go listen(clientPort, logger, dataChan)
	log.Println(<-logger)
	msg := "hello client in test"
	sendStr(outgoingPort, client.Addr, msg)
	received := <-dataChan
	actual := received.String()
	if actual != msg {
		log.Println(actual)
		t.Error("sendTo failed")
	}
}

func TestStreamerWithTwoClients(t *testing.T) {
	const (
		incomingPort = "60020"
		outgoingPort = "60011"
		clientPort1  = "60066"
		clientPort2  = "60077"

		masterPort = "60555"
	)
	output := make(chan string)
	// connect client with id 66
	go RunStreamer(incomingPort, outgoingPort, output)
	log.Println(<-output)
	log.Println(<-output)
	clientSend(clientPort1, outgoingPort, "CONNECT 66")

	log.Println(<-output)

	if Clients[Id("66")] == nil {
		t.Error("client 66 hasn't connected")
		t.FailNow()
	}
	// connect client with id 77
	clientSend(clientPort2, outgoingPort, "CONNECT 77")
	log.Println(<-output)
	// log.Println(<-output)

	if Clients[Id("77")] == nil {
		t.Error("client 77 hasn't connected")
		t.FailNow()
	}
	log.Println("len(Clients)")
	log.Println(len(Clients))
	log.Println("client connected")

	data1 := make(chan *Data)
	go listen(clientPort1, output, data1)
	data2 := make(chan *Data)
	go listen(clientPort2, output, data2)
	log.Println(<-output)
	log.Println(<-output)
	msg := "hello clients"
	clientSend(masterPort, incomingPort, msg)

	d := <-data1
	if actual := d.String(); actual != msg {
		t.Errorf("client 66 didn't receive message '%s' from Streamer, instead it received '%s'", msg, actual)
	}

	d = <-data2
	if actual := d.String(); actual != msg {
		t.Errorf("client 77 didn't receive message '%s' from Streamer, instead it received '%s'", msg, actual)
	}
}
