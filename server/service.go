package server

import (
	pb "chat/proto"
	"errors"
	"fmt"
)

type User struct {
	name string
}

func (u User) GetName() string {
	return u.name
}

func NewUser(name string) *User {
	return &User{
		name: name,
	}
}

type Topic struct {
	name        string
	topicType   pb.Topic_Type
	subscribers map[string]*User
}

func (t *Topic) GetName() string {
	return t.name
}

func (t *Topic) Subscribe(user *User) {
	t.subscribers[user.name] = user
}

func (t *Topic) Unsubscribe(user *User) {
	delete(t.subscribers, user.name)
}

func (t *Topic) IsEmptySubscribers() bool {
	return len(t.subscribers) == 0
}

func (t *Topic) IsUserSubscribed(user *User) bool {
	_, ok := t.subscribers[user.name]
	return ok
}

func (t *Topic) IsGroup() bool {
	return t.topicType == pb.Topic_GROUP
}

func (t *Topic) Convert() *pb.Topic {
	return &pb.Topic{Title: t.name, Type: t.topicType}
}

func (t *Topic) GetSubscribers() map[string]*User {
	return t.subscribers
}

func NewTopic(name string, topicType pb.Topic_Type) *Topic {
	return &Topic{
		name:        name,
		topicType:   topicType,
		subscribers: make(map[string]*User),
	}
}

type Storage interface {
	GetTopic(topic string) *Topic
	SaveTopic(topic *Topic) error
	CreateTopic(topic *Topic) error
	GetTopics() []*Topic
	Delete(topic *Topic) error
}

type PubSub interface {
	CreateChannel(username string) <-chan *pb.Msg
	GetChannel(username string) chan *pb.Msg
	DeleteChannel(username string)
}

type Service struct {
	storage Storage
	pubsub  PubSub
}

func NewService() *Service {
	storage := NewLocalStorage()
	pubsub := NewLocalPubSub()
	return &Service{
		storage: &storage,
		pubsub:  &pubsub,
	}
}

func (p *Service) Connect(username string) (<-chan *pb.Msg, error) {
	if t := p.storage.GetTopic(username); t != nil {
		return nil, errors.New(fmt.Sprintf("this username %s already exists", username))
	}
	t := NewTopic(username, pb.Topic_PERSONAL)
	if err := p.storage.CreateTopic(t); err != nil {
		return nil, err
	}
	return p.pubsub.CreateChannel(username), nil
}

func (p *Service) JoinGroupChat(topic, username string) error {
	if t := p.storage.GetTopic(topic); t != nil {
		t.Subscribe(&User{name: username})
		return p.storage.SaveTopic(t)

	}
	return errors.New(fmt.Sprintf("group chat \"%s\" does not exist", topic))
}

func (p *Service) LeftGroupChat(topic, username string) error {
	if t := p.storage.GetTopic(topic); t != nil {
		t.Unsubscribe(&User{name: username})
		if err := p.storage.SaveTopic(t); err != nil {
			return err
		}
		if t.IsEmptySubscribers() {
			return p.storage.Delete(t)
		}
		return nil
	}
	return errors.New(fmt.Sprintf("group chat \"%s\" does not exist", topic))
}

func (p *Service) CreateTopic(topic, username string) error {
	if t := p.storage.GetTopic(topic); t != nil {
		return errors.New(fmt.Sprintf("chat name \"%s\" already taken", topic))
	}
	t := NewTopic(topic, pb.Topic_GROUP)
	t.Subscribe(NewUser(username))
	return p.storage.CreateTopic(t)
}

func (p *Service) SendMessage(topic, text, username string) error {
	t := p.storage.GetTopic(topic)
	if t == nil {
		return errors.New(fmt.Sprintf("group chat with name %s does not exist", topic))
	}

	if t.IsGroup() && !t.IsUserSubscribed(&User{name: username}) {
		return errors.New(fmt.Sprintf("you are not joined to the group chat: \"%v\"", topic))
	}

	msg := &pb.Msg{Text: text, Author: username, Topic: t.Convert()}

	if !t.IsGroup() {
		p.pubsub.GetChannel(t.GetName()) <- msg
		return nil
	}

	for _, user := range t.GetSubscribers() {
		if username == user.GetName() {
			continue
		}
		p.pubsub.GetChannel(user.GetName()) <- msg
	}
	return nil
}

func (p *Service) GetTopics() ([]*pb.Topic, error) {
	ts := p.storage.GetTopics()
	topics := make([]*pb.Topic, len(ts))
	for _, tpc := range ts {
		topics = append(topics, tpc.Convert())
	}
	return topics, nil
}

func (p *Service) UnsubscribeAll(username string) {
	topics := p.storage.GetTopics()

	for _, t := range topics {
		t.Unsubscribe(&User{name: username})
		if t.IsGroup() && t.IsEmptySubscribers() {
			p.storage.Delete(t)
		} else {
			p.storage.SaveTopic(t)
		}

	}

	p.storage.Delete(&Topic{name: username})

	p.pubsub.DeleteChannel(username)
	return
}
