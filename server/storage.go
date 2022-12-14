package server

import (
	"errors"
	"fmt"
	"sync"
)

type LocalStorage struct {
	topics map[string]*Topic
	mut    sync.RWMutex
}

func NewLocalStorage() LocalStorage {
	return LocalStorage{
		topics: make(map[string]*Topic),
	}
}

func (s *LocalStorage) GetTopic(topic string) *Topic {
	t, ok := s.topics[topic]
	if !ok {
		return nil
	}
	return t
}

func (s *LocalStorage) SaveTopic(topic *Topic) error {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.topics[topic.GetName()] = topic
	return nil
}

func (s *LocalStorage) GetTopics() []*Topic {
	v := make([]*Topic, 0, len(s.topics))

	for _, value := range s.topics {
		v = append(v, value)
	}

	return v
}

func (s *LocalStorage) Delete(topic *Topic) error {
	delete(s.topics, topic.GetName())
	return nil
}

func (s *LocalStorage) CreateTopic(topic *Topic) error {
	s.mut.Lock()
	defer s.mut.Unlock()
	if _, ok := s.topics[topic.GetName()]; ok != false {
		return errors.New(fmt.Sprintf("chat name \"%s\" already taken", topic.GetName()))
	}

	s.topics[topic.GetName()] = topic
	return nil
}
