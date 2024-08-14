package core

import (
	"context"
	"time"
)

// Event type constants.
const (
	EventTypeBookmarkAdded              = "bookmark:added"
	EventTypeBookmarkNameChanged        = "bookmark:name_changed"
	EventTypeBookmarkDescriptionChanged = "bookmark:description_changed"
	EventTypeBookmarkUrlChanged         = "bookmark:url_changed"
	EventTypeBookmarkRemoved            = "bookmark:removed"
)

// Event represents an event that occurs in the system. These events are
// eventually propagated out to connected users via WebSockets whenever changes
// occur so that the UI can update in real-time.
type Event struct {
	// Specifies the type of event that is occurring.
	Type string `json:"type"`

	// The actual data from the event. See related payload types below.
	Payload interface{} `json:"payload"`
}

// BookmarkNameChangedPayload represents the payload for an Event object with a
// type of EventTypeBookmarkNameChanged.
type EventTypeBookmarkAddedPayload = Event

type EventTypeBookmarkNameChangedPayload struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type EventTypeBookmarkDescriptionChangedPayload struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type EventTypeBookmarkUrlChangedPayload struct {
	ID        int       `json:"id"`
	Url       string    `json:"url"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type EventTypeBookmarkRemovedPayload struct {
	ID int `json:"id"`
}

// EventService represents a service for managing event dispatch and event
// listeners (aka subscriptions).
//
// Events are user-centric in this implementation although a more generic
// implementation can use a topic-centic model. This application of event
// subscriptions does not makse sense for the intended use cases.
// The application has frequent reconnects so it's more efficient to subscribe
// for a single user instead of resubscribing to all their related topics.
type EventService interface {
	// Publishes an event to a user's event listeners.
	// If the user is not currently subscribed then this is a no-op.
	PublishEvent(userID string, event Event)

	// Creates a subscription for the current user's events.
	// Caller must call Subscription.Close() when done with the subscription.
	Subscribe(ctx context.Context) (Subscription, error)
}

// NopEventService returns an event service that does nothing.
func NopEventService() EventService { return &nopEventService{} }

type nopEventService struct{}

func (*nopEventService) PublishEvent(userID string, event Event) {}

func (*nopEventService) Subscribe(ctx context.Context) (Subscription, error) {
	panic("not implemented")
}

// Subscription represents a stream of events for a single user.
type Subscription interface {
	// Event stream for all user's event.
	C() <-chan Event

	// Closes the event stream channel and disconnects from the event service.
	Close() error
}
