package main

import (
	"net"
	"fmt"
	"bufio"
	"os"
)

const (
	PORT = "10050"
	PREFIX_NICKNAME = "NICKNAME"
)

type MessageType int
const (
	NicknamePrompt MessageType = iota+1
	NewMessage
	ServerDisconnected
)

type Message struct {
	Type MessageType
	Text string
}

func receiveMessages(conn net.Conn, messages chan Message) {
	buf := make([]byte,1024)
	for {
		len, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error: ",err)
			return
		}
		msg := string(buf[0:len])
		if msg == PREFIX_NICKNAME {
			messages <- Message{
				Type: NicknamePrompt,
			}
			continue
		}
		fmt.Println(msg)
	}
}

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:"+PORT) //string(PORT)
    if err != nil {
        fmt.Println("Error: ", err)
        return
    }
    defer conn.Close()

	messages := make(chan Message)

	go receiveMessages(conn, messages)

	//wait for prompt to enter nickname
	for {
		msg := <- messages
		if msg.Type == NicknamePrompt {
			fmt.Println("Welcome to the chat! Please enter a nickname...")
			reader := bufio.NewReader(os.Stdin)
			nickname, _ := reader.ReadString('\n')
			conn.Write([]byte(PREFIX_NICKNAME+nickname))
			break
		}
	}

	//wait for and send new message
	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		conn.Write([]byte(text))
	}
}