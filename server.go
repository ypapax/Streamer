package main

import (
	"fmt"
	"log"
	"net"
	"runtime"
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
	go incomingServer(incomingPort, output)
	go outgoingServer(outgoingPort, output)
}

func outgoingServer(outgoingPort string, logger chan string) {
	dataChan := make(chan *Data)
	go listen(outgoingPort, logger, dataChan)

	for {
		data := <-dataChan
		outgoingClientMsg := data.String()
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
			log.Println("client.LastAliveTime before alive command", client.LastAliveTime)
			client.LastAliveTime = time.Now()
			log.Println("client.LastAliveTime after alive command", client.LastAliveTime)

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
	log.Println("checking timeout for client", client.Id)
	if client.LastAliveTime.Before(now) {
		delete(Clients, client.Id)
		msg := fmt.Sprint("client ", client.Id, " was removed because he didn't send ALIVE command more than ", DISCONNECT_TIMEOUT)
		log.Println(msg)
		output <- msg
	}
}

type Data struct {
	Bytes []byte
	N     int
	Addr  *net.UDPAddr
}

func (d *Data) String() string {
	return string(d.Bytes[:d.N])
}

func listen(port string, logger chan string, received chan *Data) {
	/* Lets prepare a address at any address at port */
	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+port)
	CheckError(err)

	log.Println("going to listen from port ", port)
	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()

	buf := make([]byte, 1024)
	logger <- fmt.Sprint("udp port is listening: ", port)
	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		CheckError(err)
		log.Printf("%s:%d gets %s from %s:%d\n", ServerAddr.IP, ServerAddr.Port, string(buf[:n]), addr.IP, addr.Port)
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
		// log.Println("getting incoming message: ", string)
		log.Println("len(Clients)", len(Clients))
		for _, client := range Clients {
			sendTo("0", client, buf)
		}
	}

}

func sendTo(outgoingPort string, client *Client, data *Data) {
	msg := fmt.Sprint("sending outgoing message for client ", client.Id)
	log.Println(msg)

	send(outgoingPort, client.Addr, data)
}

func send(outgoingPort string, targetAddr *net.UDPAddr, data *Data) {
	sendBytes(outgoingPort, targetAddr, data.Bytes)
}

func sendBytes(outgoingPort string, remoteAddr *net.UDPAddr, buf []byte) {

	localAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+outgoingPort)
	CheckError(err)

	log.Printf("going to send %s from %s:%d to %s:%d\n", string(buf), localAddr.IP, localAddr.Port, remoteAddr.IP, remoteAddr.Port)

	Conn, err := net.DialUDP("udp", localAddr, remoteAddr)
	CheckError(err)

	log.Printf("sending %s from %s:%d to %s:%d\n", string(buf), localAddr.IP, localAddr.Port, remoteAddr.IP, remoteAddr.Port)

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
	log.Println("stack", stack)
}
