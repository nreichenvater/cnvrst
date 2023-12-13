package main

import (
	"fmt"
	"net"
	"strings"
	"time"
	//"io"
	//"os"
	//"unicode"
)

const (
	PORT = "10050"
	PREFIX_NICKNAME = "HFtgBh2Kqf8Gfpkl6N2Coskw8i6qHO0D"
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
			s := fmt.Sprintf("%s [%s]: %s",now.Format(time.RFC3339),msg.Client.Nickname,msg.Text)
			fmt.Println(s)
			for _,client := range clients {
				if client.Conn.RemoteAddr().String() != addr {
					client.Conn.Write([]byte(msg.Text))
				}
			}
		}
	}
}

/*func prepareNickname(input string) string {
	s := ""
	for _,c := range input {
		if unicode.IsPrint(c) { //necessary to remove blank runes that produce new lines...
			s = s + string(c)
		}
	}
	return s
}*/

func registerClient(conn net.Conn, messages chan Message) {
	defer conn.Close()
	buff := make([]byte,256)
	messages <- Message{
		Type: ClientConnected,
		Client: Client{Conn:conn},
	}
	cli := Client{
		Conn: conn,
		Nickname: "",
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