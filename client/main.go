package main

import (
	"net"
	"fmt"
	"strings"
	"github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"
)

const (
	PORT = 10050
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

func receiveMessages(conn net.Conn, messages chan Message, textView *tview.TextView) {
	buf := make([]byte,256)
	for {
		text := ""
		messageComplete := false
		for ; messageComplete == false ; {
			ln, err := conn.Read(buf)
			if err != nil {
				fmt.Println("Error: ",err)
				return
			}
			text = text + string(buf[0:ln])
			if strings.HasSuffix(text,PROTOCOL_SUFFIX) { messageComplete = true }
			text = strings.TrimRight(text, PROTOCOL_SUFFIX)
			if text == PREFIX_NICKNAME {
				messages <- Message{
					Type: NicknamePrompt,
				}
			} else {
				fmt.Fprintf(textView, "\n%s", tview.Escape(text))
				textView.ScrollToEnd()
			}
		}
	}
}

func getWelcomePageFlex(pages *tview.Pages, conn net.Conn) (*tview.Flex, *tview.InputField) {
	welcomeHeadingFlex := tview.NewFlex().
	AddItem(nil, 0, 1, false).
	AddItem(tview.NewTextView().SetText("Welcome to the CNVRSTE chatroom!"), 0, 1, false).
	AddItem(nil, 0, 1, false)

	nicknameInputField := tview.NewInputField().
		SetLabel("Please enter a nickname (1-20 characters): ").
		SetFieldWidth(20)

	nicknameInputField.SetDoneFunc(func (key tcell.Key){
			input := nicknameInputField.GetText()
			if len := len(input); len < 1 || len > 20 {
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

func getChatPageFlex(pages *tview.Pages, textView *tview.TextView) (*tview.Flex, *tview.InputField) {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	textView.SetDynamicColors(true).SetBorder(true)
	input := tview.NewInputField()
	flex.AddItem(textView, 0, 15, false)
	flex.AddItem(input, 0, 1, true)
	return flex, input	
}

func main() {
	connstr := fmt.Sprintf("%s%d","127.0.0.1:",PORT)
	conn, err := net.Dial("tcp",connstr)
    if err != nil {
        fmt.Println("Error: ", err)
        return
    }
    defer conn.Close()

	messages := make(chan Message)

	app := tview.NewApplication().EnableMouse(true)
	textView := tview.NewTextView().SetChangedFunc(func() {
		app.Draw()
	}).SetScrollable(true)

	go receiveMessages(conn,messages,textView)

	pages := tview.NewPages()
	
	welcomePageFlex, nicknameInputField := getWelcomePageFlex(pages, conn)
	pages.AddPage("welcome", welcomePageFlex, true, true)

	
	chatPageFlex, chatInputField := getChatPageFlex(pages,textView)
	pages.AddPage("chat", chatPageFlex, true, false)

	chatInputField.SetDoneFunc(func (key tcell.Key){
		input := chatInputField.GetText()
		if len(input) < 1 { return }
		text := fmt.Sprintf("%s%s",input,PROTOCOL_SUFFIX)
		conn.Write([]byte(text))
		chatInputField.SetText("")
	})

	for {
		msg := <- messages
		switch msg.Type {
		case NicknamePrompt:
			if err := app.SetRoot(pages, true).SetFocus(nicknameInputField).Run(); err != nil {
				panic(err)
			}
		}
	}
}