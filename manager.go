package box

import (
	"context"
	"sync"
)

type Manager struct {
	subscribers []Subscriber
}

func NewManager() (*Manager, error) {
	return &Manager{
		subscribers: make([]Subscriber, 0),
	}, nil
}

func (im *Manager) AddProcessors(processors ...Subscriber) {
	im.subscribers = append(im.subscribers, processors...)
}

func (im *Manager) Run(ctx context.Context) {
	var wg sync.WaitGroup

	for _, c := range im.subscribers {
		wg.Add(1)
		go func(s Subscriber) {
			err := s.Start(ctx)
			if err != nil {
				panic(err)
			}

			defer wg.Done()
		}(c)
	}

	wg.Wait()
}
