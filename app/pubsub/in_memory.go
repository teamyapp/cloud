package pubsub

import (
	"sync"
)

type InMemory struct {
	mutex         sync.Mutex
	subscriptions map[string][]*InMemorySubscription
}

var _ PubSub = (*InMemory)(nil)

func (i *InMemory) Subscribe(topic string, callback func(data interface{})) Subscription {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	subscription := &InMemorySubscription{
		pubSub:   i,
		topic:    topic,
		callback: callback,
	}
	i.subscriptions[topic] = append(i.subscriptions[topic], subscription)
	return subscription
}

func (i *InMemory) Publish(topic string, data interface{}) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	subs, ok := i.subscriptions[topic]
	if !ok {
		return nil
	}

	for _, sub := range subs {
		go func(sub *InMemorySubscription) {
			sub.callback(data)
		}(sub)
	}

	return nil
}

func NewInMemory() *InMemory {
	return &InMemory{
		subscriptions: make(map[string][]*InMemorySubscription),
	}
}

type InMemorySubscription struct {
	pubSub   *InMemory
	topic    string
	callback func(data interface{})
}

var _ Subscription = (*InMemorySubscription)(nil)

func (i *InMemorySubscription) Unsubscribe() error {
	i.pubSub.mutex.Lock()
	defer i.pubSub.mutex.Unlock()

	subs := i.pubSub.subscriptions[i.topic]
	newSubs := make([]*InMemorySubscription, 0)

	for _, sub := range subs {
		if sub == i {
			continue
		}
		newSubs = append(newSubs, sub)
	}

	if len(newSubs) == 0 {
		delete(i.pubSub.subscriptions, i.topic)
	} else {
		i.pubSub.subscriptions[i.topic] = newSubs
	}
	return nil
}
