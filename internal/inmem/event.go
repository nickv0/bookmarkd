package inmem

import (
	"context"
	"fmt"
	"sync"

	"bookmarkd"
	"bookmarkd/internal/core"
)

// EventBufferSize is the buffer size of the channel for each subscription.
const EventBufferSize = 16

// Ensure type implements interface.
var _ core.EventService = (*EventService)(nil)

// EventService represents a service for managing events in the system.
type EventService struct {
	mu sync.Mutex
	m  map[string]map[*Subscription]struct{} // subscriptions by user ID
}

// NewEventService returns a new instance of EventService.
func NewEventService() *EventService {
	return &EventService{
		m: make(map[string]map[*Subscription]struct{}),
	}
}

// PublishEvent publishes event to all of a user's subscriptions.
//
// If user's channel is full then the user is disconnected. This is to prevent
// slow users from blocking progress.
func (s *EventService) PublishEvent(userID string, event core.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Skip if the user is not subscribed at all.
	subs := s.m[userID]
	if len(subs) == 0 {
		return
	}

	// Publish event to all subscriptions for the user.
	for sub := range subs {
		select {
		case sub.c <- event:
		default:
			s.unsubscribe(sub)
		}
	}
}

// Subscribe creates a new subscription for the currently logged in user.
// Returns EUNAUTHORIZED if user is not logged in.
func (s *EventService) Subscribe(ctx context.Context) (core.Subscription, error) {
	// Fetch current user's ID.
	userID := core.GetUserIDFromContext(ctx)
	if userID == "" {
		return nil, fmt.Errorf("must be logged in to subscribe to events: %w", bookmarkd.ErrUnauthorized)
	}

	// Create new subscription for the user.
	sub := &Subscription{
		service: s,
		userID:  userID,
		c:       make(chan core.Event, EventBufferSize),
	}

	// Add to list of user's subscriptions.
	// Subscritions are stored as a map for each user so we can easily delete them.
	subs, ok := s.m[userID]
	if !ok {
		subs = make(map[*Subscription]struct{})
		s.m[userID] = subs
	}
	subs[sub] = struct{}{}

	return sub, nil
}

// Unsubscribe disconnects sub from the service.
func (s *EventService) Unsubscribe(sub *Subscription) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.unsubscribe(sub)
}

func (s *EventService) unsubscribe(sub *Subscription) {
	// Only close the underlying channel once. Otherwise Go will panic.
	sub.once.Do(func() {
		close(sub.c)
	})

	// Find subscription map for user. Exit if one does not exist.
	subs, ok := s.m[sub.userID]
	if !ok {
		return
	}

	// Remove subscription from map.
	delete(subs, sub)

	// Stop tracking user if they no longer have any subscriptions.
	if len(subs) == 0 {
		delete(s.m, sub.userID)
	}
}

// Ensure type implements interface.
var _ core.Subscription = (*Subscription)(nil)

// Subscription represents a stream of user-related events.
type Subscription struct {
	service *EventService // service subscription was created from
	userID  string        // subscribed user

	c    chan core.Event // channel of events
	once sync.Once       // ensures c only closed once
}

// Close disconnects the subscription from the service it was created from.
func (s *Subscription) Close() error {
	s.service.Unsubscribe(s)
	return nil
}

// C returns a receive-only channel of user-related events.
func (s *Subscription) C() <-chan core.Event {
	return s.c
}
