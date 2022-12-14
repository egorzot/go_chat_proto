package server

import pb "chat/proto"

type LocalPubSub struct {
	subscribers map[string]*Subscriber
}

func NewLocalPubSub() LocalPubSub {
	return LocalPubSub{
		subscribers: make(map[string]*Subscriber),
	}
}

type Subscriber struct {
	c chan *pb.Msg
}

func NewSubscriber(name string) *Subscriber {
	return &Subscriber{
		c: make(chan *pb.Msg),
	}
}

func (s *Subscriber) Delete() {
	close(s.c)
}

func (s *Subscriber) GetChannel() chan *pb.Msg {
	return s.c
}

func (p LocalPubSub) CreateChannel(username string) <-chan *pb.Msg {
	p.subscribers[username] = NewSubscriber(username)
	return p.subscribers[username].GetChannel()
}

func (p LocalPubSub) GetChannel(username string) chan *pb.Msg {
	return p.subscribers[username].GetChannel()
}

func (p LocalPubSub) DeleteChannel(username string) {
	p.subscribers[username].Delete()
	delete(p.subscribers, username)
}
