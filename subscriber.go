package box

import (
	"context"
)

// Subscriber defines the methods for a message subscriber.
type Subscriber interface {
	// Start initiates the subscription to receive and process messages. This method
	// begins the message retrieval and handling loop. It should be designed to
	// continuously poll or listen for messages and dispatch them to the registered handlers.
	// Returns an error if the subscription process cannot be started.
	Start(ctx context.Context) error

	// RegisterHandler associates a message type with a handler. This method is used to
	// specify how different types of messages should be processed. The messageType parameter
	// is a string that identifies the type of message, and the handler is the corresponding
	// Handler that will process those messages. Returns an error if the handler cannot be registered.
	RegisterHandler(messageType string, handler Handler)

	// UnregisterHandler removes the association of a message type with a handler.
	// This can be used to dynamically change how messages are processed during runtime.
	// The messageType parameter is a string that identifies the type of message.
	// Returns an error if the handler cannot be unregistered.
	UnregisterHandler(messageType string) error

	// Status returns the current status of the Subscriber, which can include information like
	// whether it is currently running, the number of processed messages, etc.
	Status() SubscriberStatus
}

// Define the SubscriberStatus struct
type SubscriberStatus struct {
	ProcessedMessageCount int
	// Other relevant status fields
}
