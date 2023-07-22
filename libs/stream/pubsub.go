package stream

import (
	"sync"
)

type Subscription[Item any] struct {
	pubSub *PubSub[Item]
	output chan Item
}

func (s *Subscription[Item]) Unsubscribe() {
	s.pubSub.unsubscribe(s)
	close(s.output)
}

func (s *Subscription[Item]) Output() <-chan Item {
	return s.output
}

func newSubscription[Item any](pubSub *PubSub[Item]) *Subscription[Item] {
	return &Subscription[Item]{
		pubSub: pubSub,
		output: make(chan Item),
	}
}

type PubSub[Item any] struct {
	subscriptions   map[*Subscription[Item]]bool
	subscriptionsMu sync.RWMutex
}

func (p *PubSub[Item]) unsubscribe(subscription *Subscription[Item]) {
	p.subscriptionsMu.Lock()
	defer p.subscriptionsMu.Unlock()
	delete(p.subscriptions, subscription)
}

func (p *PubSub[Item]) Subscribe() *Subscription[Item] {
	subscription := newSubscription[Item](p)
	p.subscriptionsMu.Lock()
	defer p.subscriptionsMu.Unlock()
	p.subscriptions[subscription] = true
	return subscription
}

func NewPubSub[Item any](input <-chan Item) *PubSub[Item] {
	pubSub := &PubSub[Item]{
		subscriptions: make(map[*Subscription[Item]]bool),
	}
	go func() {
		for item := range input {
			pubSub.subscriptionsMu.RLock()
			subscriptions := copyMap[*Subscription[Item], bool](pubSub.subscriptions)
			pubSub.subscriptionsMu.RUnlock()
			for subscription := range subscriptions {
				subscription.output <- item
			}
		}

		pubSub.subscriptionsMu.RLock()
		subscriptions := copyMap[*Subscription[Item], bool](pubSub.subscriptions)
		pubSub.subscriptionsMu.RUnlock()
		for subscription := range subscriptions {
			subscription.Unsubscribe()
		}
	}()
	return pubSub
}

func copyMap[Key comparable, Value any](m map[Key]Value) map[Key]Value {
	copiedMap := make(map[Key]Value)
	for k, v := range m {
		copiedMap[k] = v
	}
	
	return copiedMap
}
