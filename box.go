package box

import (
	"context"
	"encoding/json"
	"time"

	"github.com/thefabric-io/transactional"
)

// Message defines the structure and behavior of messages within the system.
type Message interface {
	// Offset returns the position of the message in the stream or queue.
	Offset() int

	// ID returns a unique identifier for the message, useful for tracking and deduplication.
	ID() string

	// Type returns the type of the message, used to determine the appropriate handler
	Type() string

	// CreatedAt returns the timestamp when the message was created.
	CreatedAt() time.Time

	// RegisteredAt returns the timestamp when the message was registered in the system.
	RegisteredAt() time.Time

	// Payload returns the actual data of the message in a byte array format.
	Payload() json.RawMessage

	// Metadata returns additional data about the message, not part of the core payload.
	Metadata() json.RawMessage
}

// Box is an interface that abstracts the operations for storing and retrieving messages.
type Box interface {
	// Init initializes the Box. This method is used to perform any setup operations
	Init(ctx context.Context, transaction transactional.Transaction) error

	DefaultLogFields() map[string]any

	// Exists checks if a message with the given ID exists in the Box. This method is used
	// to check if a message exists in the Box. It returns a boolean value indicating
	// whether the message exists or not.
	Exists(ctx context.Context, transaction transactional.Transaction, id string) (bool, error)

	// UpdateConsumer updates the offset of the consumer. This method is used to update the
	// offset (acked and consumed) of the consumer.
	UpdateConsumer(ctx context.Context, transaction transactional.Transaction, name string, offsetAcked, offsetConsumed int) error

	// Receive takes a message and stores it in the Box. This method is used to add messages
	// to the Box, typically by producers or publishers of messages.
	// It returns an error if the message cannot be stored.
	Receive(ctx context.Context, transaction transactional.Transaction, msg Message) error

	// Retrieve fetches a batch of messages from the Box for processing. It returns a slice
	// of Messages and an error. If there are no messages available, it waits for a
	// specified duration before returning. The batch size and wait time can be configured
	// to optimize performance based on specific use cases.
	Retrieve(ctx context.Context, transaction transactional.Transaction, consumer string, onlyTypes []string, batchSize int, waitTime time.Duration) ([]Message, error)
}
