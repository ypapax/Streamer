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

func outgoingServer(outgoingPort string, output chan string) {
	/* Lets prepare a address at any address at port */
	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+outgoingPort)
	CheckError(err)

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()

	buf := make([]byte, 1024)

	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		CheckError(err)

		outgoingClientMsg := string(buf[0:n])
		log.Println(outgoingClientMsg)
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
			go client.CheckTimoutLater(output)
			break
		case "CONNECT":
			client := &Client{
				Id:            id,
				Addr:          addr,
				LastAliveTime: time.Now(),
			}
			Clients[id] = client
			go client.CheckTimoutLater(output)
			output <- fmt.Sprintf("client with id %s connected", id)
			break
		case "DISCONNECT":
			delete(Clients, id)
			output <- fmt.Sprintf("client with id %s disconnected", id)
			break
		}
	}
}

func (client *Client) CheckTimoutLater(output chan string) {
	now := time.Now()
	time.Sleep(DISCONNECT_TIMEOUT)
	if Clients[client.Id] == nil {
		return
	}
	log.Println("checking timeout for client", client.Id)
	if client.LastAliveTime.Before(now) {
		delete(Clients, client.Id)
		msg := fmt.Sprint("client ", client.Id, " was removed because he didn't send KEEP ALIVE command more than ", DISCONNECT_TIMEOUT)
		log.Println(msg)
		output <- msg
	}
}
