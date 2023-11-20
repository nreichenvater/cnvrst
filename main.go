package main

import (
	"fmt"
	"net"
)

const (
	PORT = "10050"
)

type MessageType int
const (
	ClientConnected MessageType = iota+1
	ClientDisconnected
	NewMessage 
)

type Message struct {
	Type MessageType
	Text string
	Conn net.Conn
}

/*
	TODO
		- protokoll überlegen wenn username kommt (trennzeichenkette, start, ende)
		- senden: farbe von user + timestamp + text
*/

func runServer(messages chan Message) {
	clients := map[string]net.Conn{}
	for {
		msg := <- messages
		addr := msg.Conn.RemoteAddr().String()
		switch msg.Type {
		case ClientConnected:
			clients[addr] = msg.Conn
			fmt.Println("Client connected from "+addr+"...")
		case ClientDisconnected:
			delete(clients,addr)
			fmt.Println("Client from "+addr+" left...")
		case NewMessage:
			fmt.Printf(msg.Text)
			for _,conn := range clients {
				if conn.RemoteAddr().String() != addr {
					conn.Write([]byte(msg.Text))
				}
			}
		}
	}
	
}

func registerClient(conn net.Conn, messages chan Message) {
	messages <- Message{
		Type: ClientConnected,
		Conn: conn,
	}
	defer conn.Close()
	buff := make([]byte,256)
	for {
		len, err := conn.Read(buff)
		if err != nil {
			messages <- Message{
				Type: ClientDisconnected,
				Conn: conn,
			}
			return
		}
		messages <- Message{
			Type: NewMessage,
			Conn: conn,
			Text: string(buff[0:len]),
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