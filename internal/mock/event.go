package mock

import (
	"context"

	"bookmarkd/internal/core"
)

var _ core.EventService = (*EventService)(nil)

type EventService struct {
	PublishEventFn func(userID string, event core.Event)
	SubscribeFn    func(ctx context.Context) (core.Subscription, error)
}

func (s *EventService) PublishEvent(userID string, event core.Event) {
	s.PublishEventFn(userID, event)
}

func (s *EventService) Subscribe(ctx context.Context) (core.Subscription, error) {
	return s.SubscribeFn(ctx)
}

type Subscription struct {
	CloseFn func() error
	CFn     func() <-chan core.Event
}

func (s *Subscription) Close() error {
	return s.CloseFn()
}

func (s *Subscription) C() <-chan core.Event {
	return s.CFn()
}
