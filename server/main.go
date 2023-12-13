package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	PORT = "10050"
	PREFIX_NICKNAME = "HFtgBh2Kqf8Gfpkl6N2Coskw8i6qHO0D"
	PROTOCOL_SUFFIX = "\r\n\r\n"
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

func runServer(messages chan Message) {
	clients := map[string]*Client{}
	for {
		msg := <- messages
		addr := msg.Client.Conn.RemoteAddr().String()
		switch msg.Type {
		case ClientConnected:
			s := fmt.Sprintf("User joined from %s...",addr)
			fmt.Println(s)
			msg.Client.Conn.Write([]byte(PREFIX_NICKNAME))
		case ClientJoinedChat:
			clients[addr] = &msg.Client
			fmt.Println(msg.Client.Nickname)
			str := fmt.Sprintf("%s joined the chat",msg.Client.Nickname)
			fmt.Println(str)
		case ClientDisconnected:
			delete(clients,addr)
			s := fmt.Sprintf("%s left...",msg.Client.Nickname)
			fmt.Println(s)
			fmt.Println(msg.Client.Nickname)
		case NewMessage:
			now := time.Now()
			timestamp := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", 
				now.Year(), now.Month(), now.Day(), 
				now.Hour(), now.Minute(), now.Second()) 
			s := fmt.Sprintf("%s [%s]: %s",timestamp,msg.Client.Nickname,msg.Text)
			fmt.Println(s)
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
	buff := make([]byte,32)
	messages <- Message{
		Type: ClientConnected,
		Client: Client{Conn:conn},
	}
	cli := Client{
		Conn: conn,
		Nickname: "",
	}
	for {
		text := ""
		messageComplete := false
		for ; messageComplete == false ; {
			ln, err := conn.Read(buff)
			if err != nil || ln < 1 {
				messages <- Message{
					Type: ClientDisconnected,
					Client: Client{Conn:conn},
				}
				return
			}
			text = text + string(buff[0:ln])
			if strings.HasSuffix(text,PROTOCOL_SUFFIX) { messageComplete = true }
			text = strings.TrimRight(text, "\r\n")
		}
		if strings.HasPrefix(text, PREFIX_NICKNAME) {
			_, nickname, _ := strings.Cut(text, PREFIX_NICKNAME)
			cli.Nickname = nickname
			messages <- Message{
				Type: ClientJoinedChat,
				Client: cli,
			}
		} else {
			messages <- Message{
				Type: NewMessage,
				Client: cli,
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