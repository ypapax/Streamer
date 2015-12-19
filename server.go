package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

type (
	Id string

	Client struct {
		Id            Id
		Addr          *net.UDPAddr
		LastAliveTime time.Time
	}
)

var (
	Clients                          = make(map[Id]*Client)
	DISCONNECT_TIMEOUT time.Duration = 30 * time.Second
)

func RunStreamer(incomingPort, outgoingPort string, output chan string) {
	go incomingServer(incomingPort, outgoingPort, output)
	go outgoingServer(outgoingPort, output)
}

func outgoingServer(outgoingPort string, output chan string) {

	/* Lets prepare a address at any address at port */
	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+outgoingPort)
	CheckError(err)

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()
	output <- fmt.Sprint("starting outgoing Server with port", outgoingPort)
	buf := make([]byte, 1024)

	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		CheckError(err)

		outgoingClientMsg := string(buf[0:n])
		log.Println("outgoingClientMsg", outgoingClientMsg)
		msgParts := strings.Split(outgoingClientMsg, " ")
		const allowedWordsInMsg = 2
		if len(msgParts) != allowedWordsInMsg {
			continue
		}
		command := msgParts[0]
		id := Id(msgParts[1])
		// log.Println("id", id)
		switch command {
		case "ALIVE":
			if Clients[id] == nil {
				continue
			}
			client := Clients[id]
			client.LastAliveTime = time.Now()
			go client.checkTimoutLater(output)
			break
		case "CONNECT":
			client := &Client{
				Id:            id,
				Addr:          addr,
				LastAliveTime: time.Now(),
			}
			Clients[id] = client
			go client.checkTimoutLater(output)
			output <- fmt.Sprintf("client with id %s connected", id)
			break
		case "DISCONNECT":
			delete(Clients, id)
			output <- fmt.Sprintf("client with id %s disconnected", id)
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
	log.Println("checking timeout for client", client.Id)
	if client.LastAliveTime.Before(now) {
		delete(Clients, client.Id)
		msg := fmt.Sprint("client ", client.Id, " was removed because he didn't send ALIVE command more than ", DISCONNECT_TIMEOUT)
		log.Println(msg)
		output <- msg
	}
}

func incomingServer(incomingPort, outgoingPort string, output chan string) {
	/* Lets prepare a address at any address at port */
	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+incomingPort)
	CheckError(err)

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()

	buf := make([]byte, 1024)

	for {
		n, _, err := ServerConn.ReadFromUDP(buf)
		log.Println("getting incoming message with bytes count: ", n)
		CheckError(err)
		log.Println("len(Clients)", len(Clients))
		for _, client := range Clients {
			sendTo(client, outgoingPort, buf)
		}
	}
}

func sendTo(client *Client, outgoingPort string, buf []byte) {
	log.Println("11111111111111111111111111 ", client.Id)
	msg := fmt.Sprint("sending outgoing message for client ", client.Id)
	log.Println(msg)
	log.Println("11111111111111111111111111 ", client.Id)

	// output <- msg

	log.Println("1111111111111", msg)
	ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+outgoingPort)
	CheckError(err)

	// LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+"10012")
	// CheckError(err)

	Conn, err := net.DialUDP("udp", ServerAddr, client.Addr)
	CheckError(err)

	defer Conn.Close()
	log.Println("1111111111111", msg)
	_, err = Conn.Write(buf)
	CheckError(err)
}
