package main

import (
	"net"
	"fmt"
	"bufio"
	"os"
	"strings"
)

const (
	PORT = "10050"
	PREFIX_NICKNAME = "HFtgBh2Kqf8Gfpkl6N2Coskw8i6qHO0D"
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
	conn, err := net.Dial("tcp", "127.0.0.1:"+PORT)
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
			input, _ := reader.ReadString('\n')
			nickname := fmt.Sprintf("%s%s",PREFIX_NICKNAME,input)
			nickname = strings.TrimRight(nickname, "\r\n")
			conn.Write([]byte(nickname))
			break
		}
	}

	//wait for and send new message
	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.TrimRight(text, "\r\n")
		conn.Write([]byte(text))
	}
}