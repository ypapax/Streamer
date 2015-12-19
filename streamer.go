package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"runtime"
	"strings"
	"time"
)

func main() {
	incomingPort := flag.String("incoming-port", "33333", "The server listens for an incoming stream on a port defined by a command line argument `incoming-port`. Any packet received on this port is immediately sent to any connected client.")
	outgoingPort := flag.String("outgoing-port", "44444", "The server listens for client connections on a port defined by a command line argument `outgoing-port`.")
	flag.Parse()
	log.Println("incoming-port", *incomingPort)
	log.Println("outgoing-port", *outgoingPort)
	output := make(chan string)
	RunStreamer(*incomingPort, *outgoingPort, output)
	for {
		log.Println(<-output)
	}
}

type (
	Id     string
	Client struct {
		Id            Id
		Addr          *net.UDPAddr
		LastAliveTime time.Time
	}
	Data struct {
		Bytes []byte
		N     int
		Addr  *net.UDPAddr
	}
)

var (
	Clients                          = make(map[Id]*Client)
	DISCONNECT_TIMEOUT time.Duration = 30 * time.Second
)

func RunStreamer(incomingPort, outgoingPort string, output chan string) {
	go incomingServer(incomingPort, output)
	go outgoingServer(outgoingPort, output)
}

func outgoingServer(outgoingPort string, logger chan string) {
	dataChan := make(chan *Data)
	go listen(outgoingPort, logger, dataChan)

	for {
		data := <-dataChan
		outgoingClientMsg := data.String()
		msgParts := strings.Split(outgoingClientMsg, " ")
		const allowedWordsInMsg = 2
		if len(msgParts) != allowedWordsInMsg {
			continue
		}
		command := msgParts[0]
		id := Id(msgParts[1])
		switch command {
		case "ALIVE":
			if Clients[id] == nil {
				continue
			}
			client := Clients[id]
			client.LastAliveTime = time.Now()
			go client.checkTimoutLater(logger)
			break
		case "CONNECT":
			client := &Client{
				Id:            id,
				Addr:          data.Addr,
				LastAliveTime: time.Now(),
			}
			Clients[id] = client
			go client.checkTimoutLater(logger)
			logger <- fmt.Sprintf("client with id %s connected", id)
			break
		case "DISCONNECT":
			delete(Clients, id)
			logger <- fmt.Sprintf("client with id %s disconnected", id)
			break
		}
	}
}

func (client *Client) checkTimoutLater(output chan string) {
	now := time.Now()
	time.Sleep(DISCONNECT_TIMEOUT)
	if Clients[client.Id] == nil {
		return
	}
	if client.LastAliveTime.Before(now) {
		delete(Clients, client.Id)
		msg := fmt.Sprint("client ", client.Id, " was removed because he didn't send ALIVE command more than ", DISCONNECT_TIMEOUT)
		output <- msg
	}
}

func (d *Data) String() string {
	return string(d.Bytes[:d.N])
}

func listen(port string, logger chan string, received chan *Data) {
	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+port)
	CheckError(err)

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()

	buf := make([]byte, 1024)
	logger <- fmt.Sprint("udp port is listening: ", port)
	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		CheckError(err)
		received <- &Data{
			Bytes: buf[:n],
			N:     n,
			Addr:  addr,
		}
	}
}

func incomingServer(incomingPort string, logger chan string) {
	incomingPackages := make(chan *Data)
	go listen(incomingPort, logger, incomingPackages)
	for {
		buf := <-incomingPackages
		for _, client := range Clients {
			// send via any free port
			sendTo("0", client, buf)
		}
	}
}

func sendTo(outgoingPort string, client *Client, data *Data) {
	send(outgoingPort, client.Addr, data)
}

func send(outgoingPort string, targetAddr *net.UDPAddr, data *Data) {
	sendBytes(outgoingPort, targetAddr, data.Bytes)
}

func sendBytes(outgoingPort string, remoteAddr *net.UDPAddr, buf []byte) {
	localAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+outgoingPort)
	CheckError(err)
	Conn, err := net.DialUDP("udp", localAddr, remoteAddr)
	CheckError(err)
	defer Conn.Close()
	_, err = Conn.Write(buf)
	CheckError(err)
}

func sendStr(outgoingPort string, targetAddr *net.UDPAddr, msg string) {
	sendBytes(outgoingPort, targetAddr, []byte(msg))
}

func CheckError(err error) {
	if err != nil {
		printStack()
		log.Fatal("Error: ", err)
	}
}

func printStack() {
	b := make([]byte, 2048)
	n := runtime.Stack(b, false)
	stack := string(b[:n])
	log.Println(stack)
}
