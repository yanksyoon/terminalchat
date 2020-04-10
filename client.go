package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	terminalchat "github.com/charlie4284/terminalchat/proto"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc"
)

var termSem = semaphore.NewWeighted(1)

func main() {
	fmt.Println("Set your username and press Enter")
	reader := bufio.NewReader(os.Stdin)
	username, err := reader.ReadString('\n')

	client, err := connect(strings.TrimSpace(username))
	if err != nil {
		panic(err)
	}

	inputChan := make(chan string)
	go getinput(inputChan)

	for {
		select {
		case msg := <-inputChan:
			broadcastmsg := terminalchat.Message{
				Username: username,
				Message:  msg,
			}
			_, err := (*client).Broadcast(context.Background(), &broadcastmsg)
			if err != nil {
				fmt.Println("54:", err)
			}
		}
	}
}

func connect(username string) (*terminalchat.TerminalchatClient, error) {
	conn, err := grpc.Dial("127.0.0.1:9999", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := terminalchat.NewTerminalchatClient(conn)
	cli, err := client.Join(context.Background(), &terminalchat.User{Username: username})
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			msg, err := cli.Recv()
			if err != nil {
				panic(err)
			}
			fmt.Println(msg.GetMessage())
		}
	}()
	return &client, nil
}

func getinput(msgChan chan string) {
	for {
		inputReader := bufio.NewReader(os.Stdin)
		msg, err := inputReader.ReadString('\n')
		if err != nil {
			fmt.Println(":84", err)
			break
		}
		msgChan <- strings.TrimSpace(msg)
	}
}
