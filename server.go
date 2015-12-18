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

	go func() {
		for {
			now := time.Now()
			log.Println("sleeping for", DISCONNECT_TIMEOUT)
			time.Sleep(DISCONNECT_TIMEOUT)
			log.Println("awaik")

			for id, client := range Clients {
				log.Println("checking client with id", id)
				if client.LastAliveTime.Before(now) {
					delete(Clients, id)
					msg := fmt.Sprint("client with id ", id, " was removed because he didn't send KEEP ALIVE command more than ", DISCONNECT_TIMEOUT)
					log.Println(msg)
					output <- msg
				}
			}
		}
	}()

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
			Clients[id].LastAliveTime = time.Now()
			break
		case "CONNECT":
			Clients[id] = &Client{
				Id:            id,
				Addr:          addr,
				LastAliveTime: time.Now(),
			}
			output <- fmt.Sprintf("client with id %s connected", id)
			break
		case "DISCONNECT":
			delete(Clients, id)
			output <- fmt.Sprintf("client with id %s disconnected", id)
			break
		}
	}
}
