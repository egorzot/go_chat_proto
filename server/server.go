package server

import (
	pb "chat/proto"
	"context"
	"errors"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
)

const (
	metadataUsername = "username"
)

type chatService interface {
	Connect(username string) (<-chan *pb.Msg, error)
	JoinGroupChat(topic, username string) error
	LeftGroupChat(topic, username string) error
	CreateTopic(topic, username string) error
	SendMessage(topic, text, username string) error
	GetTopics() ([]*pb.Topic, error)
	UnsubscribeAll(username string)
}

type Server struct {
	pb.UnimplementedChatServer
	chat chatService
}

func NewServer() *Server {
	return &Server{
		chat: NewService(),
	}
}

func (s *Server) Connect(in *pb.ConnectRequest, srv pb.Chat_ConnectServer) error {
	if in.GetUsername() == "" {
		return errors.New("username must not be empty")
	}

	sb, err := s.chat.Connect(in.GetUsername())
	if err != nil {
		return err
	}

	logrus.Infoln("Connected a new user: " + in.GetUsername())

	for {

		select {
		case <-srv.Context().Done():
			go s.chat.UnsubscribeAll(in.GetUsername())
			logrus.Infoln(fmt.Sprintf("User %s disconnected", in.GetUsername()))
			return nil
		case msg, ok := <-sb:
			if !ok {
				logrus.Infoln(fmt.Sprintf("Subscriber channel for %s is closed. exit", in.GetUsername()))
				break
			}
			if err := srv.Send(msg); err != nil {
				logrus.Errorln("Error while sending message to stream: " + err.Error())
			}
		}
	}
}

func (s *Server) JoinGroupChat(ctx context.Context, request *pb.JoinGroupRequest) (*empty.Empty, error) {
	username := metadata.ValueFromIncomingContext(ctx, metadataUsername)[0]
	if username == "" {
		return &empty.Empty{}, errors.New("username is not set")
	}

	if request.GetTopic() == "" {
		return &empty.Empty{}, errors.New("group name must not be empty")
	}

	if err := s.chat.JoinGroupChat(request.GetTopic(), username); err != nil {
		return &empty.Empty{}, err
	}

	logrus.Infoln(fmt.Sprintf("User %s joined group %s", username, request.GetTopic()))

	return &empty.Empty{}, nil
}

func (s *Server) LeftGroupChat(ctx context.Context, request *pb.LeftGroupRequest) (*empty.Empty, error) {
	username := metadata.ValueFromIncomingContext(ctx, metadataUsername)[0]

	if username == "" {
		return &empty.Empty{}, errors.New("username is not defined")
	}

	if request.GetTopic() == "" {
		return &empty.Empty{}, errors.New("group name must not be empty")
	}

	if err := s.chat.LeftGroupChat(request.GetTopic(), username); err != nil {
		return &empty.Empty{}, errors.New(err.Error())
	}

	logrus.Infoln(fmt.Sprintf("User %s left group %s", username, request.GetTopic()))

	return &empty.Empty{}, nil
}

func (s *Server) CreateGroupChat(ctx context.Context, request *pb.CreateGroupRequest) (*empty.Empty, error) {
	username := metadata.ValueFromIncomingContext(ctx, metadataUsername)[0]

	if username == "" {
		return &empty.Empty{}, errors.New("username is not defined")
	}

	if request.GetTopic() == "" {
		return &empty.Empty{}, errors.New("group name must not be empty")
	}

	if err := s.chat.CreateTopic(request.GetTopic(), username); err != nil {
		return &empty.Empty{}, errors.New(err.Error())
	}

	logrus.Infoln(fmt.Sprintf("User %s created and automatically joined group %s", username, request.GetTopic()))

	return &empty.Empty{}, nil
}

func (s *Server) SendMessage(ctx context.Context, request *pb.SendMessageRequest) (*empty.Empty, error) {
	username := metadata.ValueFromIncomingContext(ctx, metadataUsername)[0]

	if username == "" {
		return &empty.Empty{}, errors.New("username is not defined")
	}

	if request.GetTopic() == "" {
		return &empty.Empty{}, errors.New("recipient must not be empty")
	}

	if username == request.GetTopic() {
		return &empty.Empty{}, errors.New("you can't send message yourself")
	}

	if request.GetText() == "" {
		return &empty.Empty{}, errors.New("message text must not be empty")
	}

	if err := s.chat.SendMessage(request.GetTopic(), request.GetText(), username); err != nil {
		return &empty.Empty{}, errors.New(err.Error())
	}
	logrus.Infoln(fmt.Sprintf("User %s sent message to %s", username, request.GetTopic()))

	return &empty.Empty{}, nil
}

func (s *Server) ListChannels(context.Context, *empty.Empty) (*pb.ListChannelsResponse, error) {
	topics, err := s.chat.GetTopics()
	if err != nil {
		return nil, err
	}
	return &pb.ListChannelsResponse{Topics: topics}, nil
}
