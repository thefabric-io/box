package box

import (
	"context"

	"github.com/thefabric-io/transactional"
)

// Handler is responsible for processing messages.
type Handler interface {
	// HandleEvent processes a single message. It takes a context for cancellation and timeouts,
	// a transaction object for handling transactional logic, and the message to be processed.
	// It returns an error if the message cannot be processed successfully.
	HandleEvent(ctx context.Context, tx transactional.Transaction, msg Message) error
}
