package main

import (
	"net"
	"fmt"
	"bufio"
	"os"
	"strings"
	"github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"
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

func getWelcomePageFlex(pages *tview.Pages, conn net.Conn) (*tview.Flex, *tview.InputField) {
	welcomeHeadingFlex := tview.NewFlex().
	AddItem(nil, 0, 1, false).
	AddItem(tview.NewTextView().SetText("Welcome to the CNVRSTE chatroom!"), 0, 1, false).
	AddItem(nil, 0, 1, false)

	nicknameInputField := tview.NewInputField().
		SetLabel("Please enter a nickname: ").
		SetFieldWidth(20)

	nicknameInputField.SetDoneFunc(func (key tcell.Key){
			input := nicknameInputField.GetText()
			if len := len(input); len < 1 || len > 20 {
				//fmt.Println("The nickname must have a length between 1 and 20 characters...")
				return
			}
			text := fmt.Sprintf("%s%s%s",PREFIX_NICKNAME,input,PROTOCOL_SUFFIX)
			conn.Write([]byte(text))
			pages.SwitchToPage("chat")
		})
		

	welcomeNicknameFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(nicknameInputField, 0, 4, true).
		AddItem(nil, 0, 1, false)

	welcomePageFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(welcomeHeadingFlex, 0, 1, false).
			AddItem(welcomeNicknameFlex, 0, 1, false).
			AddItem(nil, 0, 1, false), 0, 2, false).
		AddItem(nil, 0, 1, false)
	
	return welcomePageFlex, nicknameInputField
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


	app := tview.NewApplication()
	pages := tview.NewPages()
        
	chatPageFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewTextView().SetText("chat"), 0, 1, false).
		AddItem(nil, 0, 1, false)
	
	welcomePageFlex, nicknameInputField := getWelcomePageFlex(pages, conn)
	pages.AddPage("Welcome", welcomePageFlex, true, true)
	pages.AddPage("chat", chatPageFlex, true, false)

	//wait for prompt to enter nickname, then show page
	for {
		msg := <- messages
		if msg.Type == NicknamePrompt {
			if err := app.SetRoot(pages, true).SetFocus(nicknameInputField).Run(); err != nil {
				panic(err)
			}
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