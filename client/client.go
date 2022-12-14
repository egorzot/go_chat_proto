package client

import (
	"bufio"
	pb "chat/proto"
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/metadata"
	"log"
	"os"
	"strings"
)

type client struct {
}

func NewChatClient() client {
	return client{}
}

var reader = bufio.NewReader(os.Stdin)

func (c client) StartChat(chatClient pb.ChatClient) {
	ctx := context.Background()

	fmt.Println("Input the username and press enter:")
	username := readString()

	cc, err := chatClient.Connect(ctx, &pb.ConnectRequest{Username: username})
	if err != nil {
		log.Fatalf(err.Error())
	}

	authorCtx := metadata.NewOutgoingContext(
		ctx,
		metadata.Pairs("username", username),
	)

	go func() {
		for {
			msg, err := cc.Recv()

			if err != nil {
				log.Fatalf("Server disconected")
			}

			if msg.GetTopic().GetType() == pb.Topic_PERSONAL {
				fmt.Printf("New message. Text: %s. Author: %s\n", msg.GetText(), msg.GetAuthor())
			} else {
				fmt.Printf("New message. Text: %s. Group: %s. Author: %s\n", msg.GetText(), msg.GetTopic().Title, msg.GetAuthor())
			}
		}
	}()

	handleMenu(chatClient, authorCtx)
}

func handleMenu(chatClient pb.ChatClient, authorCtx context.Context) {
	var inpt string
	for {
		showMenu()
		inpt = readString()
		switch inpt {
		case "1":
			fmt.Println("Channels list:")
			res, err := chatClient.ListChannels(authorCtx, &empty.Empty{})
			if err != nil {
				fmt.Println(err.Error())
			} else {
				//todo: wtf ???
				for _, v := range res.GetTopics() {
					if v.GetTitle() == "" {
						continue
					}
					t := "personal"
					if v.GetType() == pb.Topic_GROUP {
						t = "group"
					}
					fmt.Printf("name: %s, type: %s \n", v.GetTitle(), t)
				}
				fmt.Print("\n")
			}

		case "2":
			fmt.Println("Please, enter the message recipient:")
			receiver := readString()
			fmt.Println("Please, enter the message:")
			msg := readString()
			if _, err := chatClient.SendMessage(authorCtx, &pb.SendMessageRequest{Topic: receiver, Text: msg}); err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Print("\nMessage is sent successfully\n\n")
			}
		case "3":
			fmt.Println("Please, enter the group name:")
			groupName := readString()
			if _, err := chatClient.CreateGroupChat(authorCtx, &pb.CreateGroupRequest{Topic: groupName}); err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Print("\nGroup is created successfully\n\n")
			}
		case "4":
			fmt.Println("Please, enter the group chat name:")
			groupName := readString()
			if _, err := chatClient.JoinGroupChat(authorCtx, &pb.JoinGroupRequest{Topic: groupName}); err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Printf("\nSuccessfully joined group chat \"%s\"\n\n", groupName)
			}
		case "5":
			fmt.Println("Please, enter the group chat name:")
			groupName := readString()
			if _, err := chatClient.LeftGroupChat(authorCtx, &pb.LeftGroupRequest{Topic: groupName}); err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Printf("\nSuccessfully left group chat \"%s\"\n\n", groupName)
			}
		case "6":
			authorCtx.Done()
			fmt.Println("You are disconnected")
			return
		default:
		}
	}
}

func readString() string {
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("An error occurred while reading input. Please try again", err)
	}

	input = strings.TrimSuffix(input, "\n")
	return input
}

func showMenu() {
	fmt.Println(`Select the command:
1. List channels
2. Send message
3. Create group chat
4. Join group chat
5. Left group chat
6. Disconnect
`)
}
