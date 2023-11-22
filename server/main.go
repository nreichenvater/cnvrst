package main

import (
	"fmt"
	"net"
	"strings"
)

const (
	PORT = "10050"
	PREFIX_NICKNAME = "NICKNAME"
	LEN_PREFIX_NICKNAME = 8
)

type MessageType int
const (
	ClientConnected MessageType = iota+1
	ClientDisconnected
	NewMessage
	ClientJoinedChat
)

type Message struct {
	Type MessageType
	Text string
	Client Client
}

type Client struct {
	Conn net.Conn
	Nickname string
	color string
}

/*
	TODO
		- protokoll überlegen wenn username kommt (trennzeichenkette, start, ende)
		- senden: farbe von user + timestamp + text
*/

func runServer(messages chan Message) {
	clients := map[string]*Client{}
	for {
		msg := <- messages
		addr := msg.Client.Conn.RemoteAddr().String()
		switch msg.Type {
		case ClientConnected:
			fmt.Println("User joined from ")
			fmt.Printf(addr+"...")
			msg.Client.Conn.Write([]byte(PREFIX_NICKNAME))
		case ClientJoinedChat:
			clients[addr] = &msg.Client
			fmt.Println("")
			fmt.Printf(msg.Client.Nickname)
			fmt.Printf(" joined...")
		case ClientDisconnected:
			delete(clients,addr)
			fmt.Println(msg.Client.Nickname," left...")
		case NewMessage:
			fmt.Println(clients[addr].Nickname)
			fmt.Printf("\n["+clients[addr].Nickname+"]: "+msg.Text)
			for _,client := range clients {
				if client.Conn.RemoteAddr().String() != addr {
					client.Conn.Write([]byte(msg.Text))
				}
			}
		}
	}
	
}

func registerClient(conn net.Conn, messages chan Message) {
	defer conn.Close()
	buff := make([]byte,256)
	messages <- Message{
		Type: ClientConnected,
		Client: Client{Conn:conn},
	}
	for {
		ln, err := conn.Read(buff)
		if err != nil {
			messages <- Message{
				Type: ClientDisconnected,
				Client: Client{Conn:conn},
			}
			return
		}
		text := string(buff[0:ln])
		if strings.HasPrefix(text, PREFIX_NICKNAME) {
			messages <- Message{
				Type: ClientJoinedChat,
				Client: Client{
					Conn: conn,
					Nickname: string(text[LEN_PREFIX_NICKNAME:len(text)-1]),
				},
			}
		} else {
			messages <- Message{
				Type: NewMessage,
				Client: Client{Conn:conn},
				Text: text,
			}
		}
	}
}

func main() {
	listener, err := net.Listen("tcp",":"+PORT)
	if err != nil {
		fmt.Println(err)
		fmt.Println("exiting...")
		return
	}
	fmt.Printf("listening for TCP connections on port %v...", PORT)
	fmt.Println("")

	messages := make(chan Message)

	go runServer(messages)
	
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go registerClient(conn, messages)
	}
}