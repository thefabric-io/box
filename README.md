# Box - A Robust Message Processing Library for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/thefabric-io/box.svg)](https://pkg.go.dev/github.com/thefabric-io/box)
[![Go Report Card](https://goreportcard.com/badge/github.com/thefabric-io/box)](https://goreportcard.com/report/github.com/thefabric-io/box)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Box is a high-performance, transactional message processing library that provides reliable message queuing and processing capabilities using PostgreSQL as the backing store.

## Features

- 📦 PostgreSQL-backed message storage
- 🔄 FIFO (First-In-First-Out) message processing
- 🔒 Transactional guarantees
- 🎯 Message type-based routing
- 📊 Consumer offset tracking
- 🚀 Batch processing support
- 🔌 Pluggable message handlers
- 📈 Processing status monitoring

## Installation

```bash
go get github.com/thefabric-io/box
```

## Quick Start
### 1. Initialize the Box

```go
import (
    "github.com/thefabric-io/box"
    "github.com/thefabric-io/box/pgbox"
)

// Create a new PostgreSQL box
box := pgbox.NewPostgresBox("myschema", "mytopic", "mytype")
 ```

### 2. Create a Message Handler

```go
type MyHandler struct {}

func (h *MyHandler) HandleEvent(ctx context.Context, tx transactional.Transaction, msg box.Message) error {
    // Process your message here
    return nil
}
 ```

### 3. Set Up a Subscriber

```go
// Create a FIFO subscriber
subscriber := box.NewFIFOSubscriber(
    transactional,    // Your transactional interface implementation
    "myconsumer",     // Consumer name
    box,              // The box instance
    100,             // Batch size
    time.Second * 5,  // Wait time between empty batches
)

// Register message handler
subscriber.RegisterHandler("message_type", &MyHandler{})
 ```

### 4. Start Processing

```go
// Create a manager
manager, _ := box.NewManager()

// Add subscriber
manager.AddProcessors(subscriber)

// Start processing
ctx := context.Background()
manager.Run(ctx)
 ```

## Architecture

Box implements a message processing system with the following key components:

- Box : Interface for message storage and retrieval
- Message : Core message structure with metadata
- Handler : Message processing logic
- Subscriber : Message consumption and routing
- Manager : Orchestrates multiple subscribers

## Configuration

The PostgreSQL box can be configured with:

- Schema name
- Topic name
- Message type
- Batch size
- Wait time between empty batches
## Best Practices

1. Transaction Management
   
   - Always use the provided transaction context
   - Handle transaction rollbacks appropriately
2. Error Handling
   
   - Implement proper error handling in handlers
   - Use context for timeouts and cancellation
3. Message Types
   
   - Use consistent message type naming
   - Register handlers for all expected message types
4. Monitoring
   
   - Track processing status using the Status() method
   - Implement proper logging in handlers

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch ( git checkout -b feature/amazing-feature )
3. Commit your changes ( git commit -m 'Add some amazing feature' )
4. Push to the branch ( git push origin feature/amazing-feature )
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Built with sqlx
- Uses lib/pq for PostgreSQL support

## Support
For support, please open an issue in the GitHub repository.