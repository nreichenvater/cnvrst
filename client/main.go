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
	PROTOCOL_SUFFIX = "\r\n\r\n"
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
	buf := make([]byte,256)
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

func getNickname() string {
	fmt.Println("Welcome to the chat! Please enter a nickname...")
	reader := bufio.NewReader(os.Stdin)
	input := ""
	valid := false
	for ; valid == false ; {
		input, _ = reader.ReadString('\n')
		input = strings.TrimRight(input, "\r\n")
		if len := len(input); len < 1 || len > 40 {
			fmt.Println("The nickname must have a length between 1 and 40 characters...")
			
		} else {
			valid = true
		}
	}
	return fmt.Sprintf("%s%s%s",PREFIX_NICKNAME,input,PROTOCOL_SUFFIX)
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
			nickname := getNickname()
			conn.Write([]byte(nickname))
			break
		}
	}

	//wait for and send new message
	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.TrimRight(text, "\r\n")
		text = fmt.Sprintf("%s%s",text,PROTOCOL_SUFFIX)
		conn.Write([]byte(text))
	}
}